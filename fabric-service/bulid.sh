#!/usr/bin/env bash

WORKSPACE=`pwd`
FABRIC_HOME="/Users/crazywolf/go/gopath/src/github.com/hyperledger/fabric"
WEB_WOEKSPACE=`pwd`/../web-service
SDK_GO="/Users/crazywolf/go/gopath/src/github.com/hyperledger/fabric-sdk-go"

cleanAll(){
    rm -rf bin crypto-config artifacts
    cd $FABRIC_HOME
    echo "切换fabric版本为release-1.1"
    git checkout master
}


clean(){
    rm -rf crypto-config artifacts
    echo "切换fabric版本为release-1.1"
    cd $FABRIC_HOME
    git checkout -b release-1.1 origin/release-1.1
    git checkout release-1.1
    cd $WORKSPACE
}

# 编译工具
buildTool(){
    echo "buildTool >>"
    cd $FABRIC_HOME
    echo "切换fabric版本为release-1.1"
    git checkout -b release-1.1 origin/release-1.1
    git checkout release-1.1
    echo "编译configtxgen"
    make configtxgen
    echo "编译cryptogen"
    make cryptogen
    cd $WORKSPACE
    mkdir bin
    echo "复制 configtxgen 和 cryptogen 到bin目录"
    cp $FABRIC_HOME/build/bin/* ./bin
    echo "cryptogen 版本"
    ./bin/cryptogen version
    echo "configtxgen 版本"
    ./bin/configtxgen --version
}

# 生成配置文件
initConfig(){
    # 生成crypto-config.yaml 配置文件
    ./bin/cryptogen showtemplate >>  crypto-config.yaml

    # 生成configtx.yaml 和 docker-compose.yaml
    cp $FABRIC_HOME/examples/e2e_cli/configtx.yaml ./
    cp $SDK_GO/test/fixtures/dockerenv/docker-compose.yaml ./
    cp $SDK_GO/test/fixtures/dockerenv/.env ./

    cd $WEB_WOEKSPACE
    cp $SDK_GO/test/fixtures/config/config_e2e.yaml ./config.yaml
}

# 根据配置文件生成所有证书文件
createzhengshu(){
    ./bin/cryptogen generate --config=crypto-config.yaml
}

# 生成创世区块
createconfigtx(){
    mkdir artifacts
    ./bin/configtxgen --profile CitizensGenesis -outputBlock ./artifacts/orderer.genesis.block
    ./bin/configtxgen --profile CitizensChannel -outputCreateChannelTx ./artifacts/citizens.tx -channelID citizens
    ./bin/configtxgen --profile CitizensChannel -outputAnchorPeersUpdate ./artifacts/FBImspanchors.tx -channelID citizens -asOrg FBIMSP

}

clean

# buildTool

# initConfig

createzhengshu

createconfigtx




