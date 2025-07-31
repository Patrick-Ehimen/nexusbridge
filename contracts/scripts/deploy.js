const { ethers } = require("hardhat");

async function main() {
  const [deployer] = await ethers.getSigners();

  console.log("Deploying contracts with the account:", deployer.address);
  console.log(
    "Account balance:",
    ethers.formatEther(await ethers.provider.getBalance(deployer.address))
  );

  // Configuration for deployment
  const relayers = [
    "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", // Example relayer 1
    "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC", // Example relayer 2
    "0x90F79bf6EB2c4f870365E785982E1f101E93b906", // Example relayer 3
  ];
  const requiredSignatures = 2;

  // Deploy MockERC20 for testing
  console.log("\nDeploying MockERC20...");
  const MockERC20 = await ethers.getContractFactory("MockERC20");
  const mockToken = await MockERC20.deploy("Test Token", "TEST", 18);
  await mockToken.waitForDeployment();
  console.log("MockERC20 deployed to:", mockToken.target);

  // Deploy EthereumBridge
  console.log("\nDeploying EthereumBridge...");
  const EthereumBridge = await ethers.getContractFactory("EthereumBridge");
  const bridge = await EthereumBridge.deploy(
    deployer.address, // Admin
    relayers,
    requiredSignatures
  );
  await bridge.waitForDeployment();
  console.log("EthereumBridge deployed to:", bridge.target);

  // Add token support
  console.log("\nAdding token support...");
  await bridge.addSupportedToken(mockToken.target);
  console.log("Added support for MockERC20");

  // Mint some tokens to deployer for testing
  console.log("\nMinting test tokens...");
  await mockToken.mint(deployer.address, ethers.parseEther("10000"));
  console.log("Minted 10,000 TEST tokens to deployer");

  console.log("\n=== Deployment Summary ===");
  console.log("MockERC20:", mockToken.target);
  console.log("EthereumBridge:", bridge.target);
  console.log("Admin:", deployer.address);
  console.log("Relayers:", relayers);
  console.log("Required Signatures:", requiredSignatures);

  // Verify deployment
  console.log("\n=== Verification ===");
  console.log(
    "Bridge admin:",
    await bridge.hasRole(await bridge.DEFAULT_ADMIN_ROLE(), deployer.address)
  );
  console.log(
    "Token supported:",
    await bridge.isTokenSupported(mockToken.target)
  );
  console.log("Required signatures:", await bridge.requiredSignatures());
  console.log("Number of relayers:", (await bridge.getRelayers()).length);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
