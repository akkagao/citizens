rm -rf artifacts crypto-config
./bin/cryptogen generate --config=crypto-config.yaml
mkdir artifacts
./bin/configtxgen --profile CitizensChain -outputBlock ./artifacts/orderer.genesis.block
./bin/configtxgen --profile CitizensChain -outputCreateChannelTx ./artifacts/citizens.tx -channelID citizens
./bin/configtxgen --profile CitizensChain -outputAnchorPeersUpdate ./artifacts/fbiorgmspanchors.tx -channelID citizens -asOrg FBIOrg
./bin/configtxgen --profile CitizensChain -outputAnchorPeersUpdate ./artifacts/appleorgmspanchors.tx -channelID citizens -asOrg AppleOrg





bin/configtxgen -inspectBlock artifacts/orderer.genesis.block > block.txt

peer channel create -o orderer.citizens.com:7050 -c citizens -f ./citizens.tx


