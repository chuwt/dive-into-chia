# 深入理解CHIA
## before starting
```
由于最近针对chia做了一次修改，所以对chia进行了一些研究，梳理一下，共同研究

sk: 私钥
pk: 公钥，可以通过私钥获取
```

## 一起贡献
一起来打造中文共识文档
[https://docs.google.com/document/d/1e94Hd03qFUCqO4LOirR28JFV7Kf7AKgxh9eu35HafCc/edit?usp=sharing](https://docs.google.com/document/d/1e94Hd03qFUCqO4LOirR28JFV7Kf7AKgxh9eu35HafCc/edit?usp=sharing)

## keywords
```
chia-consensus
chia consensus 中文
```
## [官方共识白皮书](https://docs.google.com/document/d/1tmRIb7lgi4QfKkNaxuKOBHRmwbVlGL4f7EsBDr_5xZE/edit)
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
```
官方原图
```
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
官方图
```
![](https://lh3.googleusercontent.com/ge9Tfy3PQ9q9RqAzLS-cNglV6dhk1ep1JijGSZB9rTNDbKB6LUHMjkJ-dmYgdjkTiHvViuRaPtBIMiuZe4E3VzWkG9R-Tcb6eS4Djoz05AD8a7EyERtGOh5CTcEL1Py6pDXGLpNi)
```
- plot文件由7个表(table)组成，每个表有2**k的实体，2到7的表里每个实体有2个指针，指向前一个表的实体。
表1的每个实体有一个整数对(0-2**k, 0-2**k)，称为 x-values
- pos的证明是64个x-values的集合
- 证明从第7个表开始，从一个实体出发（具体是哪个实体是根据挑战选择的，后面会说到），通过实体的2个指针
继续查找后面的表，最终会形成两棵树（因为一个实体有两个指针），树的最低层是表1对应的x-values,
总共有64个x-values（2**7 <- 2**6 <- 2**5 <- 2**4 <- 2**3 <- 2**2 <- 2**1)
- 读取的数据量很小，官方数据每次实体的读取在10ms左右，到了表1需要640ms，因为要读64次
- 为了优化检索过程，只分别检索两棵树的1个分支（具体哪个分支取决于挑战hash），最终会得到两个x-values
- 然后计算 quality_str = hash(两个x-values), quality_str
- 然后计算 required_iters = calculate_iterations_quality(
                            self.harvester.constants.DIFFICULTY_CONSTANT_FACTOR, // 配置文件的值
                            quality_str, // 上面的quality_str
                            plot_info.prover.get_size(), // k 
                            new_challenge.difficulty, // 难度
                            new_challenge.sp_hash, // singage point hash 后面讨论
                        )
- 如果返回的required_iters 小于一定的大小（sub_slot_iters/64), 则再去获取整一个证明，
这样就会减少磁盘读取，每个分支只需要7次搜索和读取，总共需要14次，大约140ms

ps:
这里没有说challenge是怎么来的，并且是怎么在第7个表里开始的。
源码里的challenge是signage_point_challenge，由plot_id, signage_point_hash, challenge_hash计算
而来，随后通过prover.get_qualities_for_challenge(sp_challenge_hash)获取
quality_strings
```

### 过滤器 (plot filter)
```
在进行查找答案之前，会先通过一层过滤，目的是过滤掉大多plot，使他们不用扫描plot，当然也没有收益😂
具体算法是 hash(challenge+plot_id), 如果得到的字符串的前n（配置文件修改）位都是0，则他们有资格扫描plot
进行搜索答案
```

### 挖矿
由于上面的plotId有两种组成，所以存在两种挖矿方式:
```
1. farm to pool public key，通过poolPk挖矿
- 此模式下
    - plot的奖励地址(target_reward_address)必须使用poolSk进行签名
    - 未完成的区块(unfinished_block)必须通过farmerSk和localSk的多重签名才能有效
    - 代码里是先通过harvester进行localSk签名，再通过farmer进行farmerSk签名，然后进行多签验证。
      最后通过pookPk进行验证奖励地址
- 目前矿池采用这种方式，首先需要用户签名授权矿池地址（用户的poolSk签名矿池地址），然后
  在多重签名的地方做一些流程修改
- 当多机部署时，由于harvester只存储了localSk，所以即使收到攻击，也不会修改奖励地址

2. farm to pool contract puzzle hash（官方暂未开放 截止2021-04-28)
```

## 共识
```
官方图如下
```
![](https://lh4.googleusercontent.com/I_YSKpzjcwRu9-MUIKDF_ZWIoQulJD045MiVyJV7TJwjRpn5ryjx5SrH1zLBxxdjAw3Q5UK8FyXvCraqfJ5PNt04RyXg4VD0tzdCx6-dN5QE8-8a1kab-YdLW1hD8rLACIkVWWQc)
```
chia的共识依赖VDF

VDF（延迟验证函数）
    
    Verifiable 
    会生成验证，验证者不需要重新运行函数就可以验证正确性（零知识证明）
    
    Delay 
    生成会消耗真实的某段时间
    
    Function
    相同的输入，得到相同的输出，幂等

sub-slot是一系列VDF迭代的集合，他会动态调整难度，即迭代次数（sub-slot iterations)，
使整个时间维持在10min左右, r1 到 r2 就是一个sub-slot

箭头表示hash依赖，如r1的hash包含ic1的hash

c1,c2,c3 表示challenge points。在c1,c2,c3时，timelord会创建挑战（256位hash）
放入vdf，vdf会将hash迭代n次，图上是100M次，然后，timelord会将挑战和vdf结果发布给
节点。此时的消息为end-of-slot，会在每个slot结束的时候发生。

sub-slot iterations: 一个常量，即一个sub-slot的最少迭代次数

challenge：256位hash值，是farmer的挑战hash，也是challenge chain vdf的输入

共有3个平行的VDF链，每个功能各不相同
1. challenge chain
2. infused challenge chain
3. reward chain

```
### signage point 和 infusion point
```
官方图![](https://lh5.googleusercontent.com/ueZl6WRiS6JxuicpvMjLcNzY1H33-YnisVRq7-XHMm7_f6ui-v64k4AiPfhMJYIePi7Ug1SOm-uuZLq6XwX8aH7BnUWXvnKD4SHQGu6mCc17n-lZ5hXU3FKpOSJQdrzcz5y1HhBb)
```
```
challenge chain 和 reward chain （infused 没有）  的每个sub-slot被分成64个VDF
每个VDF入口被称为signage point。timelord 会在每个signage point时发布他自己的
VDF结果和证明（proof），每个signage point之间的迭代次数称为
 signage point interval iterations, 是 sub-slot-iterations / 64
 
每个sub slot开始时的challenge 也是一个有效的 signage point。每当64个signage point
都完成时，他们会被timelord和node广播出去，farmer收到之后，会根据 signage point 和 
plotId 和 sub-slot challenge 进行filter过滤，如果满足前n位（NUMBER_ZERO_BITS_PLOT_FILTER配置文件参数，官方为9）
都是0，则通过过滤。

    plot filter bits = sha256(plot id + sub slot challenge + signage point)

挑战的hash是通过filter计算得来的

    pos challenge = sha256(plot filter bits)

farmer通过pos challenge 在plot中寻找quality_strings（并非全部证明），然后通过quality_strings计算
required_iterations，最终如果required_iterations < sp_interval_iters (signage point interval_iters),
则计算整个证明，创建unfinished block，并广播

signage point iterations 是从 sub-slot 开始到现在的signage point的迭代次数

infusion iterations 是从sub-slot 开始到可以将块加入到链中的迭代次数

    infusion iterations =( signage point iterations + 3 * sp interval iterations + required iterations)  %  sub slot iterations
    
infusion iterations 一般是3-4个signage point的迭代次数。farmer必须在次之前提交他们的答案和
块，如果此时在sub-slot的末尾，会进行%，然后顺延到下一个sub-slot，称为overflow

在infusion point时，farmer的块会和infusion point VDF的输出一起作为输入，重新进行VDF。
只有当ifusion iterations 迭代完成后，并且VDF的证明到达区块后，块才有效。

上图的b1块必须包含两个VDF的证明，这样才是有效的
1. 从r1 到 signage point
2. 从r1 到 b1
比如在图中红色处产生了块，但是要在图中b1处块才会有效，需要infusion iterations VDF
当到达infusion iters后，farmer知道自己可以出块后，会获取所有证明

signage points：challenge chain 和reward chain中的 sub-slot中64个间隔点，在每个
signage point，VDF的输出会被广播到网络里，sub-slot的第一个signage point是挑战本身。
每个区块都有一个signage point，所以区块中的证明（pos）必须满足signage point的条件

required iterations：通过quality string计算的迭代次数，用来计算infusion point

infusion point：farmer产出块之后，需要在3到4个signage point的地方进行验证的点，延迟有很多
好处，比如防止孤块，自私挖矿和分叉,同时还给farmer充足的时间进行签名。
```
### 多个块
```
官方图
```
![](https://lh6.googleusercontent.com/LKGsHBj3Wy-MzVZNJa841pd632aDR6MW4zhFTAcXneewBqSVPf3XuWygznLxuMp52Sm9NCpA_67AriGRjQynpxNLlvqb1hexRHpIrU8-NavK5oCQlElhmmclX-7iq0c2ygPWGKrX)
```
多个块可以在同一个sub-slot中，系统目前是一个sub-slot包含32个块，具体多少是通过难度控制的。
每个块的VDF证明可以重叠，例如图上的B2的证明是从B1到sp2+B1到B2的VDF证明。B3包含B1到sp3和
B2到B3的证明。B2不依赖于B3，但是B3依赖B2，B3的输入是B2的输出。infusion point处不需要
签名，只需要VDF。

```
### 3条链
共有3个平行的VDF链，每个功能各不相同
1. challenge chain
2. infused challenge chain
3. reward chain
#### 缘由
```
如果我们只用一个链来运行（奖励链），则会出现通过包含和排除区块的方式控制下一个sub-slot的挑战。
攻击者可以通过组合不同的区块来找到符合他们的挑战（grinding acctacks）。
```
```
官方图
```
![https://lh4.googleusercontent.com/xP4CbQ3hiAL5a5FttPIt30Q3d_05bnViXMfb4Bi1otlNHI4ivP9jGSVBIuARRy5mI9zz__RfKyw2B9kaCSvuZ9xAkeWKj1bbY4f-oa1uW-R5GilPWJApIXSfvFfYMpthkx6TsC8j](https://lh4.googleusercontent.com/xP4CbQ3hiAL5a5FttPIt30Q3d_05bnViXMfb4Bi1otlNHI4ivP9jGSVBIuARRy5mI9zz__RfKyw2B9kaCSvuZ9xAkeWKj1bbY4f-oa1uW-R5GilPWJApIXSfvFfYMpthkx6TsC8j)
```
上图中有几个块，B1, B2, B3, B4 ...
challenge chain 和 reward chain 在整个sub-slot中会有64个signage point。每个块必须包含
这两个chain的signage point的VDF和 三条chain的infusion point的VDF
```







