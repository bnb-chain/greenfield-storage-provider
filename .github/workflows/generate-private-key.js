const ethers = require('ethers');

// Generate a random private key
const privateKey = ethers.Wallet.createRandom().privateKey;

// Create a wallet instance from the private key
const wallet = new ethers.Wallet(privateKey);

// Print the private key and account address
console.log(`Random Ethereum private key: ${privateKey}`);
console.log(`Account address: ${wallet.address}`);

// Set environment variables for the private key and account address
console.log(`::set-output name=private_key::${privateKey}`);
console.log(`::set-output name=account_address::${wallet.address}`);