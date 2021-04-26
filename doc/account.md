# 账户体系
```
chia的账户(地址)采用的是bls算法，同时支持HD钱包。
通过助记词生成主私钥 master_sk ，同时会派生几个私钥：
1. farmer_key  对应HD路径: m/12381/8444/0/0
2. pool_key    对应HD路径: m/12381/8444/1/0
3. wallet_key  对应HD路径: m/12381/8444/2/0
4. local_key   对应HD路径: m/12381/8444/3/0
5. backup_key  对应HD路径: m/12381/8444/4/0
```
## 加密算法
- [bls](https://github.com/Chia-Network/bls-signatures)
### 什么是bls
bls是一个支持多签验证的加密算法
### 
