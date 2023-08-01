# Create a SP cluster

1. Make sure cloud resources have been set up following /docs/cloud/aws.
2. Copy this folder into your workspace.
3. Make sure you have `kustomize` installed (a component of kubectl).
4. Replace the values in `kustomization.yaml` and `config.toml`.
5. You will need to create the keys. Please see [runbook](https://docs.bnbchain.org/greenfield-docs/docs/guide/storage-provider/run-book/run-testnet-SP-node) here.
6. Set up secret called `default`:

   1. Reference `secret.env`
   2. Create the secret resource directly (or set up external secret with custom CRD).

6. Run the following command to create and apply a K8S manifest:

   ```
   $ kustomize build . > sp.yaml
   $ kubectl apply -f ./sp.yaml
   ```

