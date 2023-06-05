# Authenticator

Authenticator module is used to verify users authentication. Each request arrived to SP gateway requires authentication. SP uses authentication to know who you are.

We currently abstract SP as the GfSp framework, which provides users with customizable capabilities to meet their specific requirements. Authenticator module provides an abstract interface, which is called `Authenticator`, as follows:

```go
// Authenticator is an abstract interface to verify users authentication.
type Authenticator interface {
    Modular
    // VerifyAuthentication verifies the users authentication.
    VerifyAuthentication(ctx context.Context, auth AuthOpType, account, bucket, object string) (bool, error)
    // GetAuthNonce get the auth nonce for which the dApp or client can generate EDDSA key pairs.
    GetAuthNonce(ctx context.Context, account string, domain string) (*spdb.OffChainAuthKey, error)
    // UpdateUserPublicKey updates the user public key once the dApp or client generates the EDDSA key pairs.
    UpdateUserPublicKey(ctx context.Context, account string, domain string, currentNonce int32, nonce int32,
        userPublicKey string, expiryDate int64) (bool, error)
    // VerifyOffChainSignature verifies the signature signed by user's EDDSA private key.
    VerifyOffChainSignature(ctx context.Context, account string, domain string, offChainSig string, realMsgToSign string) (bool, error)
}

// AuthOpType defines the operator type used to verify authentication.
type AuthOpType int32

const (
    // AuthOpTypeUnKnown defines the default value of AuthOpType
    AuthOpTypeUnKnown AuthOpType = iota
    // AuthOpAskCreateBucketApproval defines the AskCreateBucketApproval operator
    AuthOpAskCreateBucketApproval
    // AuthOpAskCreateObjectApproval defines the AskCreateObjectApproval operator
    AuthOpAskCreateObjectApproval
    // AuthOpTypeGetChallengePieceInfo defines the GetChallengePieceInfo operator
    AuthOpTypeGetChallengePieceInfo
    // AuthOpTypePutObject defines the PutObject operator
    AuthOpTypePutObject
    // AuthOpTypeGetObject defines the GetObject operator
    AuthOpTypeGetObject
    // AuthOpTypeGetUploadingState defines the GetUploadingState operator
    AuthOpTypeGetUploadingState
    // AuthOpTypeGetBucketQuota defines the GetBucketQuota operator
    AuthOpTypeGetBucketQuota
    // AuthOpTypeListBucketReadRecord defines the ListBucketReadRecord operator
    AuthOpTypeListBucketReadRecord
)
```

Authenticator interface inherits [Modular interface](./lifecycle_modular.md#modular-interface), so Approver module can be managed by lifycycle and resource manager.

You can overwrite `VerifyAuthentication` to implement your own authentication mode by different AuthOpType. This is the most basic authentication.

## Off-Chain Authentication

### Abstract

This document outlines an off-chain authentication specification for greenfield storage providers (SPs) and clients. The specification includes a full functional workflow and a reference implementation, making it easy for any application integrating with greenfield SPs to build an off-chain authentication mechanism.

### Motivation

Applications based on the greenfield chain often need to interact with multiple greenfield SPs, which are off-chain services that require users to use Ethereum-compatible accounts to represent their identities.

For most interactions between applications and SPs, users' identities are required. Typically, applications can use message signing via account private keys to authenticate users, as long as they have access to their private keys. However, for browser-based applications, accessing the end users' private keys directly is not possible, making it necessary to prompt users to sign messages for each off-chain request between applications and SPs. This results in a poor user experience.

This document describes a workflow to address this problem.

### WorkFlow

#### Overall workflow

![](../../../asset/015-Auth-Overview.png)

#### Step 1 - Generate EdDSA key pairs in Apps

Applications can design how to generate EdDSA key pairs themselves, and SPs do not have any restrictions on it.

Here is one example.

1. For each Ethereum address, the SP counts how many times the account public key has been updated for a given app domain since registration. This value is called the "key nonce" and denoted as n. It starts from 0 and increments by 1 after each successful account public key update. The app can invoke the [SP API "request_nonce"]() to retrieve the nonce value n. A simple request to get the nonce is:

```sh
curl --location 'https://${SP_API_ADDRESS}/auth/request_nonce' \
--header 'X-Gnfd-User-Address: 0x3d0a49B091ABF8940AD742c0139416cEB30CdEe0' \
--header 'X-Gnfd-App-Domain: https://greenfield_app1.domain.com'
```

The response is:

```json
{
    "current_nonce": 0,
    "next_nonce": 1,
    "current_public_key": "",
    "expiry_date": ""
}
```

Since we are trying to build a new key pairs, we will use next_nonce value as n.

2. The app puts `n` and `sp addresses` into a constant strings:

```plain
Sign this message to let dapp ${dapp_domain} access the following SPs:
- SP ${SP_ADDRESS_1} (name:${SP_NAME_1}) with nonce:${NONCE_1}
- SP ${SP_ADDRESS_2} (name:${SP_NAME_2}) with nonce:${NONCE_2}
...
```

We denote the new string as `M`

3. The app then requests the user to sign M with their Ethereum ECDSA private key, then gets the signature S.
4. The app uses sha256(S) as the seed to generate the EdDSA key pairs EdDSA_private_K and EdDSA_public_K.
5. The app saves EdDSA_private_K as plain text in the browser's session storage and registers EdDSA_public_K as the account public key into the SP servers.

#### Step 2 - Register EdDSA public key in SPs

For each combination of user address and app domain, the SP backend maintains a key nonce `n`. It starts from 0 and increments by 1 after each successful account key update.


To register an account public key into a certain SP, you can invoke [SP API "update\_key"](https://greenfield.bnbchain.org/docs/api-sdk/storgae-provider-rest/auth/update_key.html).

Here is an example. Suppose that

1. The **user account address** is `0x3d0a49B091ABF8940AD742c0139416cEB30CdEe0`
2. The **app domain** is `https://greenfield_app1.domain.com`
3. The **nonce** for above user address and app domain from [SP API "request\_nonce"](https://greenfield.bnbchain.org/docs/api-sdk/storgae-provider-rest/auth/get_nonce.html) is `1`
4. The **SP operator address** is `0x70d1983A9A76C8d5d80c4cC13A801dc570890819`
5. The **EdDSA\_public\_K** is `4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083`
6. The **expiry time** for this `EdDSA_public_K` is `2023-04-28T16:25:24Z`. The expiry time indicates the expiry time of this `EdDSA_public_K` , which should be a future time and within **7 days.**

The app will put above information into a text message:

```plain
https://greenfield_app1.domain.com wants you to sign in with your BNB Greenfield account:\n0x3d0a49B091ABF8940AD742c0139416cEB30CdEe0\n\nRegister your identity public key 4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083\n\nURI: https://greenfield_app1.domain.com\nVersion: 1\nChain ID: 5600\nIssued At: 2023-04-24T16:25:24Z\nExpiration Time: 2023-04-28T16:25:24Z\nResources:\n- SP 0x70d1983A9A76C8d5d80c4cC13A801dc570890819 (name: SP_001) with nonce: 1
```

We denote this text message as `M2`

and request user to sign and get the signature`S2`:

![](../../../asset/015-Auth-Update-Key-Metamask.png)

Finally, the app invokes [SP API "update\_key"](https://greenfield.bnbchain.org/docs/api-sdk/storgae-provider-rest/auth/update_key.html) by putting `S2` into http Authorization header. The following is an example:

```plain
curl --location --request POST 'https://${SP_API_ADDRESS}/auth/update_key' \
--header 'Origin: https://greenfield_app1.domain.com' \
--header 'X-Gnfd-App-Domain: https://greenfield_app1.domain.com' \
--header 'X-Gnfd-App-Reg-Nonce: 1' \
--header 'X-Gnfd-App-Reg-Public-Key: 4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083' \
--header 'X-Gnfd-App-Reg-Expiry-Date: 2023-04-28T16:25:24Z' \
--header 'Authorization: PersonalSign ECDSA-secp256k1,SignedMsg=https://greenfield_app1.domain.com wants you to sign in with your BNB Greenfield account:\n0x3d0a49B091ABF8940AD742c0139416cEB30CdEe0\n\nRegister your identity public key 4db642fe6bc2ceda2e002feb8d78dfbcb2879d8fe28e84e02b7a940bc0440083\n\nURI: https://greenfield_app1.domain.com\nVersion: 1\nChain ID: 5600\nIssued At: 2023-04-24T16:25:24Z\nExpiration Time: 2023-04-28T16:25:24Z\nResources:\n- SP 0x70d1983A9A76C8d5d80c4cC13A801dc570890819 (name: SP_001) with nonce: 1,Signature=0x8663c48cfecb611d64540d3b653f51ef226f3f674e2c390ea9ca45746b22a4f839a15576b5b4cc1051183ae9b69ac54160dc3241bbe99c695a52fe25eaf2f8c01b'
```

Once the response code returns 200, you can check if the new account public key is saved into this SP by invoking [SP API "request\_nonce"](https://greenfield.bnbchain.org/docs/api-sdk/storgae-provider-rest/auth/get_nonce.html)
This API returns the latest key nonce for a given user address and app domain.

If the API returns the new key nonce, the account public key has been successfully registered into the SP servers. The app can now use the EdDSA key pair generated in Step 1 to authenticate the user in future interactions with the SP.

#### Step 3 - Use EdDSA seed to sign request and verification

In Step1 & Step2, we generated EdDSA keys and registered them into SP. In Step3, we can use `EdDSA_private_K` to sign request when an app invokes a certain SP API.

To sign a request, the app needs to define a customized text message with a recent expiry timestamp (denoted as `EdDSA_M`) and use `EdDSA_private_K` to sign this message to get the signature `EdDSA_S`.

The text message format is `${actionContent}_${expiredTimestamp}`.

For example, if a user clicks the "download" button in an app to download a private object they own, this will invoke the SP getObject API. 
The `EdDSA_M` could be defined as `Invoke_GetObject_1682407345000`, and the `EdDSA_S` would be `a48fff140b148369a108611502acff919720b5493aa36ba0886d8d73634ee20404963b847104d06aa822cf904741aff70ede4ba7d70fa8808c3206d4c93be623`.

To combine `EdDSA_M` and `EdDSA_S`, the app should include them in the Authorization header when invoking the GetObject API:

```plain
curl --location 'https://${SP_API_ADDRESS}/${bucket_name}/${object_name}' \
--header 'authorization: OffChainAuth EDDSA,SignedMsg=Invoke_GetObject_1682407345000,Signature=a48fff140b148369a108611502acff919720b5493aa36ba0886d8d73634ee20404963b847104d06aa822cf904741aff70ede4ba7d70fa8808c3206d4c93be623' \
--header 'X-Gnfd-User-Address: 0x3d0a49B091ABF8940AD742c0139416cEB30CdEe0' \
--header 'X-Gnfd-App-Domain: https://greenfield_app1.domain.com' 
```

By including the signed message and signature in the Authorization header, the app can authenticate the request with the SP servers. The SP servers can then verify the signature using the EdDSA_public_K registered in Step 2.

#### Step 4 - Manage EdDSA key pairs

Although we defined an expiry date for registered `EdDSA_public_K`, users might want to know how many EdDSA keys they are currently using and might want to delete them for security concerns.

To list a user's registered EdDSA account public keys in an SP, apps can invoke [SP API "list\_key"](https://greenfield.bnbchain.org/docs/api-sdk/storgae-provider-rest/auth/list_key.html).
To delete a user's registered EdDSA account public key in an SP, apps can invoke  [SP API "delete\_key"](https://greenfield.bnbchain.org/docs/api-sdk/storgae-provider-rest/auth/delete_key.html)

#### Auth API Specification

See [SP Auth Rest API Doc](https://greenfield.bnbchain.org/docs/api-sdk/storgae-provider-rest/auth) 

### Rational

### Security Considerations

#### Preventing replay attacks

To prevent replay attacks, which are man-in-the-middle attacks in which an attacker is able to capture the user's signature and resend it to establish a new session for themselves, the following measures should be taken:

* A new `nonce` should be selected each time when EdDSA keys are generated. This ensures that each generated key pair is unique and cannot be replayed.

* When using EdDSA_private_K to sign a request, a recent timestamp as the expiry date must be included. This ensures that the signed message is only valid for a limited time and cannot be reused in a replay attack.

By implementing these measures, the app can minimize the risk of replay attacks and ensure the security of the user's data and interactions with the SP servers.

## GfSp Framework Authenticator Code

Authenticator module code implementation: [Authenticator](https://github.com/bnb-chain/greenfield-storage-provider/tree/master/modular/authenticator)
