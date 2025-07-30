const { ethers } = require("hardhat");

async function main() {
  console.log("Deploying NexusBridge contracts...");

  // Get the deployer account
  const [deployer] = await ethers.getSigners();
  console.log("Deploying contracts with account:", deployer.address);

  const balance = await ethers.provider.getBalance(deployer.address);
  console.log("Account balance:", ethers.formatEther(balance), "ETH");

  // TODO: Deploy bridge contracts
  console.log("Contract deployment will be implemented in future tasks");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
