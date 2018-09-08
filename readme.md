#  fabric-go-sdk 学习

## 1. 安装软件环境

此项目在macPro环境下部署

1. 安装git
2. 安装golang
3. 安装docker
4. 安装docker-compose

以上软件安装自行google吧，教程很多。

## 2. 编译工具

- 把[fabric源码](https://github.com/hyperledger/fabric)下载到 $GOPATH/src/github.com/hyperledger 目录下

  git clone git@github.com:hyperledger/fabric.git

- 切换到1.1版本（本项目使用fabirc1.1作为演示）

  git checkout -b release-1.1 origin/release-1.1

- 生成工具包

  在fabric目录$GOPATH/src/github.com/hyperledger/fabric目录下执行

  make configtxgen

  make cryptogen

  在build/bin目录下生成 configtxgen cryptogen 文件

- 把上面编译好的文件放到本项目 fabric-service/bin 目录下

```shell
➜  fabric-service git:(master) ✗ tree -l bin
bin
├── configtxgen
└── cryptogen
```

## 编写配置文件

- 约定

  本项目暂定1个组织FBI，组织有2个节点，2个用户

  根域名使用 citizens.com

### crypto-config.yaml 文件

```shell
使用下面命令生成 文件模版
./bin/cryptogen showtemplate > crypto-config.yaml
```

根据实际定义修改内容，最终结果为

```yaml
OrdererOrgs:
  - Name: Orderer
    Domain: citizens.com
    Specs:
      - Hostname: orderer
PeerOrgs:
  - Name: FBI
    Domain: fbi.citizens.com
    EnableNodeOUs: false
    Template:
      Count: 2
    Users:
      Count: 2
```

- 在fabric-service目录下生成证书目录

```
➜  fabric-service git:(master) ✗ ./bin/cryptogen generate --config=crypto-config.yaml
fbi.citizens.com
```

### configtx.yaml 文件

- 从fabric配置文件例子中获取模版

```shell
➜  fabric-service git:(master) ✗ 
cp $GOPATH/src/github.com/hyperledger/fabric/examples/e2e_cli/configtx.yaml ./
```

- 修改后的内容

```yaml
---
Profiles:

    CitizensGenesis:
        Capabilities:
            <<: *ChannelCapabilities
        Orderer:
            <<: *OrdererDefaults
            Organizations:
                - *OrdererOrg
            Capabilities:
                <<: *OrdererCapabilities
        Consortiums:
            SampleConsortium:
                Organizations:
                    - *FBI
    CitizensChannel:
        Consortium: SampleConsortium
        Application:
            <<: *ApplicationDefaults
            Organizations:
                - *FBI
            Capabilities:
                <<: *ApplicationCapabilities
Organizations:
    - &OrdererOrg
        Name: OrdererOrg
        ID: OrdererMSP
        MSPDir: crypto-config/ordererOrganizations/citizens.com/msp
    - &FBI
        Name: FBIMSP
        ID: FBIMSP
        MSPDir: crypto-config/peerOrganizations/fbi.citizens.com/msp
        AnchorPeers:
            - Host: peer0.fbi.citizens.com
              Port: 7051
Orderer: &OrdererDefaults
    OrdererType: solo
    Addresses:
        - orderer.citizens.com:7050
    BatchTimeout: 2s
    BatchSize:
        MaxMessageCount: 10
        AbsoluteMaxBytes: 98 MB
        PreferredMaxBytes: 512 KB
    Organizations:
Application: &ApplicationDefaults
    Organizations:
Capabilities:
    Global: &ChannelCapabilities
        V1_1: true
    Orderer: &OrdererCapabilities
        V1_1: true
    Application: &ApplicationCapabilities
        V1_1: true

```

- 生成order创世区块锚节点配置文件

```shell
mkdir artifacts
//生成order创世区块
./bin/configtxgen --profile CitizensGenesis -outputBlock ./artifacts/orderer.genesis.block
// 生成channel初始块
./bin/configtxgen --profile CitizensChannel -outputCreateChannelTx ./artifacts/citizens.tx -channelID citizens
//创建锚节点更新文件
./bin/configtxgen --profile CitizensChannel -outputAnchorPeersUpdate ./artifacts/FBImspanchors.tx -channelID citizens -asOrg FBIMSP
```

- 生成channel初始块

  channel 名字为 citizens

```shell
命令：
./bin/configtxgen --profile CitizensChain -outputCreateChannelTx ./artifacts/citizens.tx -channelID citizens
  ---------------------------------执行结果----------------------------------------------
➜  fabric-service git:(master) ✗ ./bin/configtxgen --profile CitizensChain -outputCreateChannelTx ./artifacts/citizens.tx -channelID citizens
2018-08-11 16:06:07.685 CST [common/tools/configtxgen] main -> INFO 001 Loading configuration
2018-08-11 16:06:07.691 CST [common/tools/configtxgen] doOutputChannelCreateTx -> INFO 002 Generating new channel configtx
2018-08-11 16:06:07.712 CST [common/tools/configtxgen] doOutputChannelCreateTx -> INFO 003 Writing new channel tx
```

- 以上命令执行完毕后查看生成的结果,如果以下文件都生成成功说明以上操作都没有问题

```shell
fabric-service git:(master) ✗ tree -l artifacts
artifacts
├── appleorgmspanchors.tx
├── citizens.tx
├── fbiorgmspanchors.tx
└── orderer.genesis.block
```

### docker-compose.yaml 文件

- 复制模版文件

```shell
基于go模版文件修改
cp $GOPATH/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/dockerenv/docker-compose.yaml ./
cp $GOPATH/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/dockerenv/.env ./
```

- 修改后的结果

```yaml
version: '2'
services:
  orderer:
    image: ${FABRIC_DOCKER_REGISTRY}${FABRIC_ORDERER_FIXTURE_IMAGE}:${FABRIC_ARCH}${FABRIC_ARCH_SEP}${FABRIC_ORDERER_FIXTURE_TAG}
    environment:
      - ORDERER_GENERAL_LOGLEVEL=info
      - ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
      - ORDERER_GENERAL_GENESISMETHOD=file
      - ORDERER_GENERAL_GENESISFILE=/etc/hyperledger/configtx/orderer.genesis.block
      - ORDERER_GENERAL_LOCALMSPID=OrdererMSP
      - ORDERER_GENERAL_LOCALMSPDIR=/etc/hyperledger/msp/orderer
      - ORDERER_GENERAL_TLS_ENABLED=true
      - ORDERER_GENERAL_TLS_PRIVATEKEY=/etc/hyperledger/tls/orderer/server.key
      - ORDERER_GENERAL_TLS_CERTIFICATE=/etc/hyperledger/tls/orderer/server.crt
      - ORDERER_GENERAL_TLS_ROOTCAS=[/etc/hyperledger/tls/orderer/ca.crt]
      - ORDERER_GENERAL_TLS_CLIENTAUTHENABLED
      - ORDERER_GENERAL_TLS_CLIENTROOTCAS
    #comment out logging.driver in order to render the debug logs
    # logging:
    #   driver: none
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric/orderer
    command: orderer
    ports:
      - 7050:7050
    volumes:
      - ./artifacts:/etc/hyperledger/configtx
      - ./crypto-config/ordererOrganizations/citizens.com/orderers/orderer.citizens.com/msp:/etc/hyperledger/msp/orderer
      - ./crypto-config/ordererOrganizations/citizens.com/orderers/orderer.citizens.com/tls:/etc/hyperledger/tls/orderer
      - ./crypto-config/peerOrganizations/fbi.citizens.com/tlsca:/etc/hyperledger/tlsca
    networks:
      default:
        aliases:
          - orderer.citizens.com

  FBIpeer1:
    image: ${FABRIC_DOCKER_REGISTRY}${FABRIC_PEER_FIXTURE_IMAGE}:${FABRIC_ARCH}${FABRIC_ARCH_SEP}${FABRIC_PEER_FIXTURE_TAG}
    environment:
      - CORE_VM_ENDPOINT
      - CORE_PEER_ID=peer0.fbi.citizens.com
      - CORE_LOGGING_PEER=info
      # - CORE_LOGGING_GRPC=debug
      # - CORE_LOGGING_GOSSIP=debug
      # - CORE_CHAINCODE_STARTUPTIMEOUT=30s
      - CORE_CHAINCODE_LOGGING_SHIM=debug
      - CORE_CHAINCODE_LOGGING_LEVEL=debug
      - CORE_CHAINCODE_BUILDER=${FABRIC_DOCKER_REGISTRY}${FABRIC_BUILDER_FIXTURE_IMAGE}:${FABRIC_ARCH}${FABRIC_ARCH_SEP}${FABRIC_BUILDER_FIXTURE_TAG}
      - CORE_CHAINCODE_GOLANG_RUNTIME=${FABRIC_BASE_DOCKER_REGISTRY}${FABRIC_BASEOS_FIXTURE_IMAGE}:${FABRIC_ARCH}${FABRIC_ARCH_SEP}${FABRIC_BASEOS_FIXTURE_TAG}
      - CORE_CHAINCODE_EXECUTETIMEOUT=120s
      - CORE_VM_DOCKER_ATTACHSTDOUT=false
      - CORE_PEER_LOCALMSPID=FBIMSP
      - CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/peer
      - CORE_PEER_LISTENADDRESS=0.0.0.0:7051
      - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:7052
      - CORE_PEER_GOSSIP_BOOTSTRAP=127.0.0.1:7051
      - CORE_PEER_ADDRESS=peer0.fbi.citizens.com:7051
      - CORE_PEER_GOSSIP_EXTERNALENDPOINT=peer0.fbi.citizens.com:7051
      - CORE_PEER_TLS_ENABLED=true
      - CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/tls/peer/server.key
      - CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/tls/peer/server.crt
      - CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/tls/peer/ca.crt
      - CORE_PEER_TLS_CLIENTAUTHREQUIRED
      - CORE_PEER_TLS_CLIENTROOTCAS_FILES
      # # the following setting starts chaincode containers on the same
      # # bridge network as the peers
      # # https://docs.docker.com/compose/networking/
      - CORE_PEER_NETWORKID=${CORE_PEER_NETWORKID}
      - CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=${CORE_PEER_NETWORKID}_default
    #comment out logging.driver in order to render the debug logs
    # logging:
    #   driver: none
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric
    command: peer node start
    ports:
      - "7051:7051"
    expose:
      - "7051"
      - "7052"
    volumes:
      - /var/run/:/var/run/
      - ./crypto-config/peerOrganizations/fbi.citizens.com/peers/peer0.fbi.citizens.com/msp:/etc/hyperledger/msp/peer
      - ./crypto-config/peerOrganizations/fbi.citizens.com/peers/peer0.fbi.citizens.com/tls:/etc/hyperledger/tls/peer
      - ./crypto-config/peerOrganizations/fbi.citizens.com/tlsca:/etc/hyperledger/orgs/fbi.citizens.com/tlsca
    networks:
      default:
        aliases:
          - peer0.fbi.citizens.com
    depends_on:
      - orderer

  FBIpeer2:
    image: ${FABRIC_DOCKER_REGISTRY}${FABRIC_PEER_FIXTURE_IMAGE}:${FABRIC_ARCH}${FABRIC_ARCH_SEP}${FABRIC_PEER_FIXTURE_TAG}
    environment:
      - CORE_VM_ENDPOINT
      - CORE_PEER_ID=peer1.fbi.citizens.com
      - CORE_LOGGING_PEER=info
      # - CORE_LOGGING_GRPC=debug
      # - CORE_LOGGING_GOSSIP=debug
      # - CORE_CHAINCODE_STARTUPTIMEOUT=30s
      - CORE_CHAINCODE_LOGGING_SHIM=debug
      - CORE_CHAINCODE_LOGGING_LEVEL=debug
      - CORE_CHAINCODE_BUILDER=${FABRIC_DOCKER_REGISTRY}${FABRIC_BUILDER_FIXTURE_IMAGE}:${FABRIC_ARCH}${FABRIC_ARCH_SEP}${FABRIC_BUILDER_FIXTURE_TAG}
      - CORE_CHAINCODE_GOLANG_RUNTIME=${FABRIC_BASE_DOCKER_REGISTRY}${FABRIC_BASEOS_FIXTURE_IMAGE}:${FABRIC_ARCH}${FABRIC_ARCH_SEP}${FABRIC_BASEOS_FIXTURE_TAG}
      - CORE_CHAINCODE_EXECUTETIMEOUT=120s
      - CORE_VM_DOCKER_ATTACHSTDOUT=false
      - CORE_PEER_LOCALMSPID=FBIMSP
      - CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/msp/peer
      - CORE_PEER_LISTENADDRESS=0.0.0.0:7151
      - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:7152
      - CORE_PEER_ADDRESS=peer1.fbi.citizens.com:7151
      - CORE_PEER_GOSSIP_EXTERNALENDPOINT=peer1.fbi.citizens.com:7151
      - CORE_PEER_GOSSIP_BOOTSTRAP=peer0.fbi.citizens.com:7051
      - CORE_PEER_TLS_ENABLED=true
      - CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/tls/peer/server.key
      - CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/tls/peer/server.crt
      - CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/tls/peer/ca.crt
      - CORE_PEER_TLS_CLIENTAUTHREQUIRED
      - CORE_PEER_TLS_CLIENTROOTCAS_FILES
      # # the following setting starts chaincode containers on the same
      # # bridge network as the peers
      # # https://docs.docker.com/compose/networking/
      - CORE_PEER_NETWORKID=${CORE_PEER_NETWORKID}
      - CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=${CORE_PEER_NETWORKID}_default
    #comment out logging.driver in order to render the debug logs
    # logging:
    #   driver: none
    working_dir: /opt/gopath/src/github.com/hyperledger/fabric
    command: peer node start
    ports:
      - "7151:7151"
    expose:
      - "7151"
      - "7152"
    volumes:
      - /var/run/:/var/run/
      - ./crypto-config/peerOrganizations/fbi.citizens.com/peers/peer1.fbi.citizens.com/msp:/etc/hyperledger/msp/peer
      - ./crypto-config/peerOrganizations/fbi.citizens.com/peers/peer1.fbi.citizens.com/tls:/etc/hyperledger/tls/peer
      - ./crypto-config/peerOrganizations/fbi.citizens.com/tlsca:/etc/hyperledger/orgs/fbi.citizens.com/tlsca
    networks:
      default:
        aliases:
          - peer1.fbi.citizens.com
    depends_on:
      - orderer

networks:
    default:

```

- 修改hosts 把以下内容加入/etc/hosts文件

```shell
127.0.0.1 peer0.fbi.citizens.com
127.0.0.1 peer1.fbi.citizens.com
127.0.0.1 apple.citizens.com
127.0.0.1 fbi.citizens.com
127.0.0.1 orderer.citizens.com
127.0.0.1 citizens.com
```

### 编写config.yaml 文件

进入web-service目录

cp $GOPATH/src/github.com/hyperledger/fabric-sdk-go/test/fixtures/config/config_e2e.yaml ./config.yaml

修改后的结果

```yaml
version: 1.0.0
client:
  organization: FBI

  logging:
    level: info
  cryptoconfig:
    path: ${GOPATH}/src/github.com/akkagao/citizens/fabric-service/crypto-config
  credentialStore:
    path: "/tmp/state-store"
    cryptoStore:
      path: /tmp/msp
  BCCSP:
    security:
     enabled: true
     default:
      provider: "SW"
     hashAlgorithm: "SHA2"
     softVerify: true
     level: 256

  tlsCerts:
    # [Optional]. Use system certificate pool when connecting to peers, orderers (for negotiating TLS) Default: false
    systemCertPool: false

    # [Optional]. Client key and cert for TLS handshake with peers and orderers
    client:
      key:
        path: 
      cert:
        path: 

channels:
  citizens:
    peers:
      peer0.fbi.citizens.com:
        endorsingPeer: true
        chaincodeQuery: true
        ledgerQuery: true
        eventSource: true
organizations:
  FBI:
    mspid: FBIMSP

    # This org's MSP store (absolute path or relative to client.cryptoconfig)
    cryptoPath:  peerOrganizations/fbi.citizens.com/users/{username}@fbi.citizens.com/msp

    peers:
      - peer0.fbi.citizens.com
      - peer1.fbi.citizens.com


    certificateAuthorities:
      - ca.fbi.citizens.com


#
orderers:
  orderer.citizens.com:
    url: localhost:7050

    # these are standard properties defined by the gRPC library
    # they will be passed in as-is to gRPC client constructor
    grpcOptions:
      ssl-target-name-override: orderer.citizens.com
      # These parameters should be set in coordination with the keepalive policy on the server,
      # as incompatible settings can result in closing of connection.
      # When duration of the 'keep-alive-time' is set to 0 or less the keep alive client parameters are disabled
      keep-alive-time: 0s
      keep-alive-timeout: 20s
      keep-alive-permit: false
      fail-fast: false
      # allow-insecure will be taken into consideration if address has no protocol defined, if true then grpc or else grpcs
      allow-insecure: false

    tlsCACerts:
      # Certificate location absolute path
      path: ${GOPATH}/src/github.com/akkagao/citizens/fabric-service/crypto-config/ordererOrganizations/citizens.com/tlsca/tlsca.citizens.com-cert.pem

#
# List of peers to send various requests to, including endorsement, query
# and event listener registration.
#
peers:
  peer0.fbi.citizens.com:
    # this URL is used to send endorsement and query requests
    url: peer0.fbi.citizens.com:7051

    grpcOptions:
      ssl-target-name-override: peer0.fbi.citizens.com
      # These parameters should be set in coordination with the keepalive policy on the server,
      # as incompatible settings can result in closing of connection.
      # When duration of the 'keep-alive-time' is set to 0 or less the keep alive client parameters are disabled
      keep-alive-time: 0s
      keep-alive-timeout: 20s
      keep-alive-permit: false
      fail-fast: false
      # allow-insecure will be taken into consideration if address has no protocol defined, if true then grpc or else grpcs
      allow-insecure: false

    tlsCACerts:
      # Certificate location absolute path
      path: ${GOPATH}/src/github.com/akkagao/citizens/fabric-service/crypto-config/peerOrganizations/fbi.citizens.com/tlsca/tlsca.fbi.citizens.com-cert.pem

  peer1.fbi.citizens.com:
    # this URL is used to send endorsement and query requests
    url: peer1.fbi.citizens.com:7151

    grpcOptions:
      ssl-target-name-override: peer1.fbi.citizens.com
      # These parameters should be set in coordination with the keepalive policy on the server,
      # as incompatible settings can result in closing of connection.
      # When duration of the 'keep-alive-time' is set to 0 or less the keep alive client parameters are disabled
      keep-alive-time: 0s
      keep-alive-timeout: 20s
      keep-alive-permit: false
      fail-fast: false
      # allow-insecure will be taken into consideration if address has no protocol defined, if true then grpc or else grpcs
      allow-insecure: false

    tlsCACerts:
      # Certificate location absolute path
      path: ${GOPATH}/src/github.com/akkagao/citizens/fabric-service/crypto-config/peerOrganizations/fbi.citizens.com/tlsca/tlsca.fbi.citizens.com-cert.pem


certificateAuthorities:
  ca.fbi.citizens.com:
    url: https://ca.fbi.citizens.com:7054
    tlsCACerts:
      # Comma-Separated list of paths
      path: ${GOPATH}/src/github.com/akkagao/citizens/fabric-service/crypto-config/peerOrganizations/fbi.citizens.com/tlsca/tlsca.fbi.citizens.com-cert.pem
      # Client key and cert for SSL handshake with Fabric CA
      client:
        key:
          path: ${GOPATH}/src/github.com/akkagao/citizens/fabric-service/crypto-config/peerOrganizations/fbi.citizens.com/users/User1@fbi.citizens.com/tls/client.key
        cert:
          path: ${GOPATH}/src/github.com/akkagao/citizens/fabric-service/crypto-config/peerOrganizations/fbi.citizens.com/users/User1@fbi.citizens.com/tls/client.crt

    # Fabric-CA supports dynamic user enrollment via REST APIs. A "root" user, a.k.a registrar, is
    # needed to enroll and invoke new users.
    registrar:
      enrollId: admin
      enrollSecret: adminpw
    # [Optional] The optional name of the CA.
    caName: ca.fbi.citizens.com
  
```

**以上就是所有配置的基本流程,也可以参考fabric-service目录下的bulid.sh脚本内容**

然后编写chainCode 和 调用链码的代码就可以测试了，具体参考chainCode和web-service目录中的go代码

## 启动体验

在fabric-service目录下执行

./start.sh  启动fabric服务，所有日志都会输出到 all.log 中

然后进入web-service 目录

使用  `go run main.go`命令启动gosdk项目 执行链码的安装、初始化、执行、查询等操作



## 以下是配置调试过程中遇到的一些错误处理方法

- 清理docker 网络

```shell
➜  fabric-service git:(master) ✗ docker network ls
NETWORK ID          NAME                DRIVER              SCOPE
c1f91c6c5086        bridge              bridge              local
aced91a76322        host                host                local
57502ba90162        none                null                local
➜  fabric-service git:(master) ✗ docker rm 57502ba90162
```

- 使用start脚本启动

  脚本会清理之前操作残留的docker 以免对当前开发环境产生影响

 如果没有报错说明启动成功，然后`docker ps `查看所有定义的容器是否都存在



1. 生成channel初始块 报错

```shell
➜  fabric-service git:(master) ✗ ./bin/configtxgen --profile CitizensChain -outputCreateChannelTx ./artifacts/citizens.tx -channelID citizens
2018-08-11 16:02:53.269 CST [common/tools/configtxgen] main -> INFO 001 Loading configuration
2018-08-11 16:02:53.276 CST [common/tools/configtxgen] doOutputChannelCreateTx -> INFO 002 Generating new channel configtx
2018-08-11 16:02:53.277 CST [common/tools/configtxgen] main -> CRIT 003 Error on outputChannelCreateTx: config update generation failure: cannot define a new channel with no Application section
➜  fabric-service git:(master) ✗
```

问题原因 configtx.yaml 文件中

```yaml
# Profiles 节点下缺少一下内容
       Application:
            <<: *ApplicationDefaults
            Organizations:
                - *FBIOrg
        Consortium: SampleConsortium
```

2. 启动docker 报错

   问题原因  docker-compose.yaml 文件 每个peer下面的CORE_PEER_LOCALMSPID 值必须正确

   - CORE_PEER_LOCALMSPID=FBIOrg

```shell
peer1.apple.citizens.com    | 2018-08-11 09:46:45.454 UTC [gossip/comm] authenticateRemotePeer -> WARN 1be Identity store rejected 192.168.0.4:7051 : failed classifying identity: Unable to extract msp.Identity from peer Identity: Peer Identity [0a 08 41 70 70 6c 65 4f 72 67 12 9e 06 2d 2d 2d 2d 2d 42 45 47 49 4e 20 43 45 52 54 49 46 49 43 41 54 45 2d 2d 2d 2d 2d 0a 4d 49 49 43 48 7a 43 43 41 63 57 67 41 77 49 42 41 67 49 51 4d 76 38 45 32 4b 67 61 42 62 61 44 5a 61 71 59 61 68 77 32 78 44 41 4b 42 67 67 71 68 6b 6a 4f 50 51 51 44 41 6a 42 33 4d 51 73 77 0a 43 51 59 44 56 51 51 47 45 77 4a 56 55 7a 45 54 4d 42 45 47 41 31 55 45 43 42 4d 4b 51 32 46 73 61 57 5a 76 63 6d 35 70 59 54 45 57 4d 42 51 47 41 31 55 45 42 78 4d 4e 55 32 46 75 49 45 5a 79 0a 59 57 35 6a 61 58 4e 6a 62 7a 45 62 4d 42 6b 47 41 31 55 45 43 68 4d 53 59 58 42 77 62 47 55 75 59 32 6c 30 61 58 70 6c 62 6e 4d 75 59 32 39 74 4d 52 34 77 48 41 59 44 56 51 51 44 45 78 56 6a 0a 59 53 35 68 63 48 42 73 5a 53 35 6a 61 58 52 70 65 6d 56 75 63 79 35 6a 62 32 30 77 48 68 63 4e 4d 54 67 77 4f 44 45 78 4d 44 55 30 4e 44 55 78 57 68 63 4e 4d 6a 67 77 4f 44 41 34 4d 44 55 30 0a 4e 44 55 78 57 6a 42 64 4d 51 73 77 43 51 59 44 56 51 51 47 45 77 4a 56 55 7a 45 54 4d 42 45 47 41 31 55 45 43 42 4d 4b 51 32 46 73 61 57 5a 76 63 6d 35 70 59 54 45 57 4d 42 51 47 41 31 55 45 0a 42 78 4d 4e 55 32 46 75 49 45 5a 79 59 57 35 6a 61 58 4e 6a 62 7a 45 68 4d 42 38 47 41 31 55 45 41 78 4d 59 63 47 56 6c 63 6a 41 75 59 58 42 77 62 47 55 75 59 32 6c 30 61 58 70 6c 62 6e 4d 75 0a 59 32 39 74 4d 46 6b 77 45 77 59 48 4b 6f 5a 49 7a 6a 30 43 41 51 59 49 4b 6f 5a 49 7a 6a 30 44 41 51 63 44 51 67 41 45 39 7a 32 79 70 37 4c 36 6b 33 58 37 76 62 43 67 70 6a 41 5a 44 76 6d 57 0a 44 46 49 61 4b 70 33 4d 36 39 72 67 61 71 44 38 51 6c 75 2b 33 68 31 4f 71 6c 32 59 4c 71 78 6b 62 6a 4d 66 76 78 52 5a 74 73 4f 65 71 57 77 41 47 42 6c 66 43 5a 78 50 6c 58 51 7a 76 71 4e 4e 0a 4d 45 73 77 44 67 59 44 56 52 30 50 41 51 48 2f 42 41 51 44 41 67 65 41 4d 41 77 47 41 31 55 64 45 77 45 42 2f 77 51 43 4d 41 41 77 4b 77 59 44 56 52 30 6a 42 43 51 77 49 6f 41 67 35 70 59 31 0a 53 36 45 6f 73 30 75 70 48 73 41 44 64 75 77 6d 45 75 4c 51 58 61 41 72 64 6b 42 4f 65 68 77 48 44 38 4b 4e 49 6d 34 77 43 67 59 49 4b 6f 5a 49 7a 6a 30 45 41 77 49 44 53 41 41 77 52 51 49 68 0a 41 49 71 4b 72 61 64 33 2f 74 61 47 38 4a 65 36 38 59 7a 38 49 78 53 37 54 49 36 6f 62 2f 37 55 36 61 72 6d 4e 48 52 6e 47 4b 70 4f 41 69 42 4d 34 32 70 73 48 4a 6a 56 73 41 4d 67 34 63 53 43 0a 36 39 38 47 51 46 61 73 74 48 56 70 68 73 6e 6f 66 38 4a 44 56 4c 53 67 56 77 3d 3d 0a 2d 2d 2d 2d 2d 45 4e 44 20 43 45 52 54 49 46 49 43 41 54 45 2d 2d 2d 2d 2d 0a] cannot be validated. No MSP found able to do that.
peer1.apple.citizens.com    | 2018-08-11 09:46:45.454 UTC [gossip/comm] Handshake -> WARN 1bf Authentication failed: failed classifying identity: Unable to extract msp.Identity from peer Identity: Peer Identity [0a 08 41 70 70 6c 65 4f 72 67 12 9e 06 2d 2d 2d 2d 2d 42 45 47 49 4e 20 43 45 52 54 49 46 49 43 41 54 45 2d 2d 2d 2d 2d 0a 4d 49 49 43 48 7a 43 43 41 63 57 67 41 77 49 42 41 67 49 51 4d 76 38 45 32 4b 67 61 42 62 61 44 5a 61 71 59 61 68 77 32 78 44 41 4b 42 67 67 71 68 6b 6a 4f 50 51 51 44 41 6a 42 33 4d 51 73 77 0a 43 51 59 44 56 51 51 47 45 77 4a 56 55 7a 45 54 4d 42 45 47 41 31 55 45 43 42 4d 4b 51 32 46 73 61 57 5a 76 63 6d 35 70 59 54 45 57 4d 42 51 47 41 31 55 45 42 78 4d 4e 55 32 46 75 49 45 5a 79 0a 59 57 35 6a 61 58 4e 6a 62 7a 45 62 4d 42 6b 47 41 31 55 45 43 68 4d 53 59 58 42 77 62 47 55 75 59 32 6c 30 61 58 70 6c 62 6e 4d 75 59 32 39 74 4d 52 34 77 48 41 59 44 56 51 51 44 45 78 56 6a 0a 59 53 35 68 63 48 42 73 5a 53 35 6a 61 58 52 70 65 6d 56 75 63 79 35 6a 62 32 30 77 48 68 63 4e 4d 54 67 77 4f 44 45 78 4d 44 55 30 4e 44 55 78 57 68 63 4e 4d 6a 67 77 4f 44 41 34 4d 44 55 30 0a 4e 44 55 78 57 6a 42 64 4d 51 73 77 43 51 59 44 56 51 51 47 45 77 4a 56 55 7a 45 54 4d 42 45 47 41 31 55 45 43 42 4d 4b 51 32 46 73 61 57 5a 76 63 6d 35 70 59 54 45 57 4d 42 51 47 41 31 55 45 0a 42 78 4d 4e 55 32 46 75 49 45 5a 79 59 57 35 6a 61 58 4e 6a 62 7a 45 68 4d 42 38 47 41 31 55 45 41 78 4d 59 63 47 56 6c 63 6a 41 75 59 58 42 77 62 47 55 75 59 32 6c 30 61 58 70 6c 62 6e 4d 75 0a 59 32 39 74 4d 46 6b 77 45 77 59 48 4b 6f 5a 49 7a 6a 30 43 41 51 59 49 4b 6f 5a 49 7a 6a 30 44 41 51 63 44 51 67 41 45 39 7a 32 79 70 37 4c 36 6b 33 58 37 76 62 43 67 70 6a 41 5a 44 76 6d 57 0a 44 46 49 61 4b 70 33 4d 36 39 72 67 61 71 44 38 51 6c 75 2b 33 68 31 4f 71 6c 32 59 4c 71 78 6b 62 6a 4d 66 76 78 52 5a 74 73 4f 65 71 57 77 41 47 42 6c 66 43 5a 78 50 6c 58 51 7a 76 71 4e 4e 0a 4d 45 73 77 44 67 59 44 56 52 30 50 41 51 48 2f 42 41 51 44 41 67 65 41 4d 41 77 47 41 31 55 64 45 77 45 42 2f 77 51 43 4d 41 41 77 4b 77 59 44 56 52 30 6a 42 43 51 77 49 6f 41 67 35 70 59 31 0a 53 36 45 6f 73 30 75 70 48 73 41 44 64 75 77 6d 45 75 4c 51 58 61 41 72 64 6b 42 4f 65 68 77 48 44 38 4b 4e 49 6d 34 77 43 67 59 49 4b 6f 5a 49 7a 6a 30 45 41 77 49 44 53 41 41 77 52 51 49 68 0a 41 49 71 4b 72 61 64 33 2f 74 61 47 38 4a 65 36 38 59 7a 38 49 78 53 37 54 49 36 6f 62 2f 37 55 36 61 72 6d 4e 48 52 6e 47 4b 70 4f 41 69 42 4d 34 32 70 73 48 4a 6a 56 73 41 4d 67 34 63 53 43 0a 36 39 38 47 51 46 61 73 74 48 56 70 68 73 6e 6f 66 38 4a 44 56 4c 53 67 56 77 3d 3d 0a 2d 2d 2d 2d 2d 45 4e 44 20 43 45 52 54 49 46 49 43 41 54 45 2d 2d 2d 2d 2d 0a] cannot be validated. No MSP found able to do that.
peer1.apple.citizens.com    | 2018-08-11 09:46:45.454 UTC [gossip/discovery] func1 -> WARN 1c0 Could not connect to {peer0.apple.citizens.com:7051 [] [] peer0.apple.citizens.com:7051 <nil>} : failed classifying identity: Unable to extract msp.Identity from peer Identity: Peer Identity [0a 08 41 70 70 6c 65 4f 72 67 12 9e 06 2d 2d 2d 2d 2d 42 45 47 49 4e 20 43 45 52 54 49 46 49 43 41 54 45 2d 2d 2d 2d 2d 0a 4d 49 49 43 48 7a 43 43 41 63 57 67 41 77 49 42 41 67 49 51 4d 76 38 45 32 4b 67 61 42 62 61 44 5a 61 71 59 61 68 77 32 78 44 41 4b 42 67 67 71 68 6b 6a 4f 50 51 51 44 41 6a 42 33 4d 51 73 77 0a 43 51 59 44 56 51 51 47 45 77 4a 56 55 7a 45 54 4d 42 45 47 41 31 55 45 43 42 4d 4b 51 32 46 73 61 57 5a 76 63 6d 35 70 59 54 45 57 4d 42 51 47 41 31 55 45 42 78 4d 4e 55 32 46 75 49 45 5a 79 0a 59 57 35 6a 61 58 4e 6a 62 7a 45 62 4d 42 6b 47 41 31 55 45 43 68 4d 53 59 58 42 77 62 47 55 75 59 32 6c 30 61 58 70 6c 62 6e 4d 75 59 32 39 74 4d 52 34 77 48 41 59 44 56 51 51 44 45 78 56 6a 0a 59 53 35 68 63 48 42 73 5a 53 35 6a 61 58 52 70 65 6d 56 75 63 79 35 6a 62 32 30 77 48 68 63 4e 4d 54 67 77 4f 44 45 78 4d 44 55 30 4e 44 55 78 57 68 63 4e 4d 6a 67 77 4f 44 41 34 4d 44 55 30 0a 4e 44 55 78 57 6a 42 64 4d 51 73 77 43 51 59 44 56 51 51 47 45 77 4a 56 55 7a 45 54 4d 42 45 47 41 31 55 45 43 42 4d 4b 51 32 46 73 61 57 5a 76 63 6d 35 70 59 54 45 57 4d 42 51 47 41 31 55 45 0a 42 78 4d 4e 55 32 46 75 49 45 5a 79 59 57 35 6a 61 58 4e 6a 62 7a 45 68 4d 42 38 47 41 31 55 45 41 78 4d 59 63 47 56 6c 63 6a 41 75 59 58 42 77 62 47 55 75 59 32 6c 30 61 58 70 6c 62 6e 4d 75 0a 59 32 39 74 4d 46 6b 77 45 77 59 48 4b 6f 5a 49 7a 6a 30 43 41 51 59 49 4b 6f 5a 49 7a 6a 30 44 41 51 63 44 51 67 41 45 39 7a 32 79 70 37 4c 36 6b 33 58 37 76 62 43 67 70 6a 41 5a 44 76 6d 57 0a 44 46 49 61 4b 70 33 4d 36 39 72 67 61 71 44 38 51 6c 75 2b 33 68 31 4f 71 6c 32 59 4c 71 78 6b 62 6a 4d 66 76 78 52 5a 74 73 4f 65 71 57 77 41 47 42 6c 66 43 5a 78 50 6c 58 51 7a 76 71 4e 4e 0a 4d 45 73 77 44 67 59 44 56 52 30 50 41 51 48 2f 42 41 51 44 41 67 65 41 4d 41 77 47 41 31 55 64 45 77 45 42 2f 77 51 43 4d 41 41 77 4b 77 59 44 56 52 30 6a 42 43 51 77 49 6f 41 67 35 70 59 31 0a 53 36 45 6f 73 30 75 70 48 73 41 44 64 75 77 6d 45 75 4c 51 58 61 41 72 64 6b 42 4f 65 68 77 48 44 38 4b 4e 49 6d 34 77 43 67 59 49 4b 6f 5a 49 7a 6a 30 45 41 77 49 44 53 41 41 77 52 51 49 68 0a 41 49 71 4b 72 61 64 33 2f 74 61 47 38 4a 65 36 38 59 7a 38 49 78 53 37 54 49 36 6f 62 2f 37 55 36 61 72 6d 4e 48 52 6e 47 4b 70 4f 41 69 42 4d 34 32 70 73 48 4a 6a 56 73 41 4d 67 34 63 53 43 0a 36 39 38 47 51 46 61 73 74 48 56 70 68 73 6e 6f 66 38 4a 44 56 4c 53 67 56 77 3d 3d 0a 2d 2d 2d 2d 2d 45 4e 44 20 43 45 52 54 49 46 49 43 41 54 45 2d 2d 2d 2d 2d 0a] cannot be validated. No MSP found able to do that.
```

3: orderer 启动失败(发现一下日志)

orderer.genesis.block 

实际创世区块名称和docker-compose-base 文件中的配置不一致导致

\- ../artifacts/genesis.block:/var/hyperledger/orderer/orderer.genesis.block

   \- ../artifacts/orderer.genesis.block:/var/hyperledger/orderer/orderer.genesis.block

```shell
orderer.ebbyte.com       | panic: Unable to bootstrap orderer. Error reading genesis block file: read /var/hyperledger/orderer/orderer.genesis.block: is a directory
orderer.ebbyte.com       |
orderer.ebbyte.com       | goroutine 1 [running]:
orderer.ebbyte.com       | github.com/hyperledger/fabric/orderer/common/bootstrap/file.(*fileBootstrapper).GenesisBlock(0xc42013c430, 0xc42013c430)
orderer.ebbyte.com       |      /opt/gopath/src/github.com/hyperledger/fabric/orderer/common/bootstrap/file/bootstrap.go:44 +0x1e4
orderer.ebbyte.com       | github.com/hyperledger/fabric/orderer/common/server.initializeBootstrapChannel(0xc4202a2a00, 0x1393300, 0xc420164000)
orderer.ebbyte.com       |      /opt/gopath/src/github.com/hyperledger/fabric/orderer/common/server/main.go:205 +0x5bd
orderer.ebbyte.com       | github.com/hyperledger/fabric/orderer/common/server.initializeMultichannelRegistrar(0xc4202a2a00, 0x138fd80, 0x13f3e20, 0xc42013e798, 0x1, 0x1, 0xc420371e60)
orderer.ebbyte.com       |      /opt/gopath/src/github.com/hyperledger/fabric/orderer/common/server/main.go:253 +0xa0
orderer.ebbyte.com       | github.com/hyperledger/fabric/orderer/common/server.Start(0xcfa0bc, 0x5, 0xc4202a2a00)
orderer.ebbyte.com       |      /opt/gopath/src/github.com/hyperledger/fabric/orderer/common/server/main.go:103 +0x24c
orderer.ebbyte.com       | github.com/hyperledger/fabric/orderer/common/server.Main()
orderer.ebbyte.com       |      /opt/gopath/src/github.com/hyperledger/fabric/orderer/common/server/main.go:82 +0x20f
orderer.ebbyte.com       | main.main()
orderer.ebbyte.com       |      /opt/gopath/src/github.com/hyperledger/fabric/orderer/main.go:15 +0x20
peer1.akka.ebbyte.com    | 2018-08-20 13:52:37.557 UTC [accessControl] newCertKeyPair -> DEBU 028 Classified peer1.akka.ebbyte.com as a hostname, adding it as a DNS SAN
```

4： ca启动失败

FABRIC_CA_SERVER_TLS_KEYFILE 配置没有修改完全 _sk 文件没有替换完

```shell
ca_peerakka              | Error: Failed to find private key for certificate in '/etc/hyperledger/fabric-ca-server-config/ca.akka.ebbyte.com-cert.pem': Could not find matching private key for SKI: Failed
getting key for SKI [[40 2 186 255 133 181 113 229 155 112 163 181 94 213 135 169 36 211 97 215 195 34 118 225 14 224 105 0 198 224 12 20]]: Key with SKI 2802baff85b571e59b70a3b55ed587a924d361d7c32276e10e
e06900c6e00c14 not found in /etc/hyperledger/fabric-ca-server/msp/keystore
```

5: go build 报错

fabric-sdk-go 版本不对 切换成 v1.0.0-alpha4 版本

```
../../../hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/orderer/configuration.pb.go:56:35: undefined: proto.InternalMessageInfo
```

6: go启动失败

OrgName 必须和configtx.yaml文件中的 Organizations定义的peer节点名称一致

```shell
Unable to initialize the Fabric SDK: failed to create channel management client from Admin identity: failed to create resmgmt client due to context error: invalid options to create identity, invalid org name
```

7: go 启动失败

config.yaml 修改为(url 值)，修改后需要重启docker

```yaml
orderers:
  orderer.ebbyte.com:
    url: 127.0.0.1:7050
```

报错信息：

```shell
Unable to initialize the Fabric SDK: failed to save channel: create channel failed: SendEnvelope failed: calling orderer 'orderer.ebbyte.com:7050' failed: Orderer Server Status Code: (400) BAD_REQUEST. Description: error authorizing update: error validating ReadSet: readset expected key [Group]  /Channel/Application at version 0, but got version 1
```

8:go启动失败

config.yaml 文件中8151 端口写错了应该是8050

go程序中所有写了 resmgmt.WithTargetEndpoints("peer0.akka.ebbyte.com")) 的地方确认域名是否是 org pee0的域名

```shell
Unable to initialize the Fabric SDK: failed to save channel: create channel failed: SendEnvelope failed: calling orderer '127.0.0.1:7050' failed: Orderer Server Status Code: (400) BAD_REQUEST. Description: error authorizing update: error validating ReadSet: readset expected key [Group]  /Channel/Application at version 0, but got version 1
```

```shell
Unable to initialize the Fabric SDK: failed to make admin join channel: join channel failed: SendProposal failed: Transaction processing for endorser [peer1.akka.ebbyte.com:8151]: Endorser Client Status Code: (2) CONNECTION_FAILED. Description: dialing connection timed out [peer1.akka.ebbyte.com:8151]
```

9 实例化ChainCode失败

docker网络问题



```shell
[fabsdk/fab] 2018/08/21 13:33:34 UTC - peer.(*peerEndorser).sendProposal -> ERRO process proposal failed [rpc error: code = Unknown desc = error starting container: API error (404): {"message":"network e2ecli_default not found"}
]
Unable to install and instantiate the chaincode: failed to instantiate the chaincode: sending deploy transaction proposal failed: Transaction processing for endorser [127.0.0.1:7051]: gRPC Transport Status Code: (2) Unknown. Description: error starting container: API error (404): {"message":"network e2ecli_default not found"}
```





重新生成证书

docker images 查看所有的iamge 如果有 network-peer0-org-domain 这样的容器用docker rmi 删除一次

确认chaincode 代码所在目录名称为chaincode 并且代码package 为main  go build 是否可以直接编译出二进制文件

```shell
[fabsdk/fab] 2018/08/21 13:46:12 UTC - peer.(*peerEndorser).sendProposal -> ERRO process proposal failed [rpc error: code = Unknown desc = error starting container: API error (400): {"message":"OCI runtime create failed: container_linux.go:348: starting container process caused \"exec: \\\"chaincode\\\": executable file not found in $PATH\": unknown"}
]
Unable to install and instantiate the chaincode: failed to instantiate the chaincode: sending deploy transaction proposal failed: Transaction processing for endorser [127.0.0.1:7051]: gRPC Transport Status Code: (2) Unknown. Description: error starting container: API error (400): {"message":"OCI runtime create failed: container_linux.go:348: starting container process caused \"exec: \\\"chaincode\\\": executable file not found in $PATH\": unknown"}
```


go:启动失败
channel Name 不能有大写字母

Unable to initialize the Fabric SDK: failed to save channel: create channel failed: SendEnvelope failed: calling orderer 'localhost:7050' failed: Orderer Server Status Code: (400) BAD_REQUEST. Description: initializing configtx manager failed: bad channel ID: channel ID 'GoldChainChannel' contains illegal characters