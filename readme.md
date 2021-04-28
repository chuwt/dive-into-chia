# 深入理解CHIA
## before starting
```
由于最近针对chia做了一次修改，所以对chia进行了一些研究，梳理一下，共同研究

sk: 私钥
pk: 公钥，可以通过私钥获取
```
## 账户
### 概述
```
chia的账户(地址)采用的是bls算法，同时支持HD钱包。
通过24个单词的助记词生成主私钥 master_sk ，同时会派生几个私钥：
1. farmer_sk  对应HD路径: m/12381/8444/0/0
2. pool_sk    对应HD路径: m/12381/8444/1/0
3. wallet_sk  对应HD路径: m/12381/8444/2/0
```
### 加密算法
- [bls](https://github.com/Chia-Network/bls-signatures) 一个支持多签验证的加密算法
### 多签流程
```
// 签名
sig1 = Sign(sk1, msg1)
sig2 = Sign(sk2, msg2)
aggSig = Aggregate(sig1, sig2)
// 验证
Verify((pk1, pk2), (msg1, msg2), aggSig)
```
## plot
### plot组成
```官方原图```
![](https://raw.githubusercontent.com/wiki/Chia-Network/chia-blockchain/images/plot_format.png)
### plotId
```
plotId有两种组成方式, 官方图如下：
```
![](https://raw.githubusercontent.com/wiki/Chia-Network/chia-blockchain/images/plot_ids.png)
```
第一种：
 - plotId 由 poolPk 和 plotPk 组成
 - plotPk 由 (localPk + farmerPk) 组成
 - poolPk 是写在plot文件中的，如果没有指定，则会使用默认（生成的第一个masterSk对应的poolPk）

第二种：
 - plotId 由 poolContract 和 plotPk 组成
 - 此模式下通过合约控制 
 
ps:
    签名区块的时候需要获取的3个key（poolPk，farmerPk，localSk）的位置
    是固定的（通过get_memo()可以获取到)，所以，理论上可以修改plot的poolPk和farmerPk
```
### 上面提到的localSk是什么
```
localSk是每个plot文件随机生成的私钥，只存在于plot文件中，使用时，会通过解析plot的头部信息，进行获取
```
### k
```
- p盘时使用 poolPk 和 famrerPk，默认使用系统的第一个masterSk派生，同时需要指定k的大小
- k控制文件大小，主网允许的最小k=32（101.4GIB, GIB是1024进制，GB是10进制)，文件大小的计算公式为 size = 780*k**2(k-10)
```

### memo
```
此处记录了 pool_public_key(48) or puzzle_hash(32), farmer_public_key(48), local_master_sk(32)
```
### plotData
```
- plot文件由7个表(table)组成，每个表有2**k的实体，每个实体有2个指针，指向前一个表，所以每个实体
有一个整数对(0-2**k, 0-2**k)，称为 x-values
- pos的证明是64个x-values的集合

```

### 挖矿
由于上面的plotId有两种组成，所以存在两种挖矿方式:
```
1. farm to pool public key，通过poolPk挖矿
- 此模式下
    - plot的奖励地址必须使用poolSk进行签名
    - 未完成的块必须通过farmerSk和localSk的多重签名才能有效
    - 代码里是先通过harvester进行localSk签名，再通过farmer进行farmerSk签名，然后进行多签验证。
      最后通过pookPk进行验证奖励地址
- 目前矿池采用这种方式，首先需要用户签名授权矿池地址（用户的poolSk签名矿池地址），然后
  在多重签名的地方做一些流程修改
- 当多机部署时，由于harvester只存储了localSk，所以即使收到攻击，也不会修改奖励地址

2. farm to pool contract puzzle hash（官方暂未开放 截止2021-04-28)
```



## 区块
## 共识