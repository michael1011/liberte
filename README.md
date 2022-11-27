# liberte

Some web3 sites like [Oasis](https://oasis.app) require Infura to work properly. But there is no reason for which you wouldn't be able to use your own node to serve those requests. Unfortunately, these sites don't let you configure the RPC endpoint they use.

By running `liberte`, doing some browser and network configuration you can "replace" Infura with your own node.

## Usage

1. Run the `gencert.sh` script in the `cert` folder and follow the instructions
2. Import the generated `myCA.pem` as certificate authority in your browser
3. Overwrite the DNS record of Infura (`mainnet.infura.io` for Ethereum mainnet) to the machine you run `liberte` on
4. Build the executable with `go build .` in the root folder of the repository
5. Run `liberte` with root permissions (because it binds to the privileged ports `80` und `443`)

You can set your node endpoint via the `--node` flag of the executable.
