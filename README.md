本项目是一个docker的精简化实现，基于go语言开发，支持容器的创建、运行、停止、查看及命令执行等核心功能，模拟了 Docker 的基本工作流程。（你问我为什么里面叫easydocker但是仓库叫minidocker，那当然是因为easy不好听啊）   

**项目概述**  
本项目实现了容器技术的核心原理，包括：  
基于namespace的进程隔离
容器镜像的拉取、存储与管理  
容器生命周期管理 
容器网络配置
简易的文件系统隔离  

注意：本项目中使用的部分外部包中的函数仅在Linux环境下可用  

**使用方法**  
1.初始化环境  
```bash
  sudo easydocker init
```
2.运行容器
```bash
  sudo easydocker run [--name containername] [--image imagename] [--it] command
```
3.查看容器列表
```bash
  sudo easydocker ps
```
4.停止容器
```bash
  sudo easydocker stop containerid
```
5.执行命令
```bash
  sudo easydocker exec containerid command
```

**项目结构**
```plaintext
easydocker/
├── cli/                # 命令行交互模块
│   ├── cli.go          # 初始化命令行应用
│   └── command/        # 具体命令实现
├── container/          # 容器管理模块
│   ├── container.go    # 容器创建、运行、停止逻辑
│   └── manager.go      # 容器信息管理
├── image/              # 镜像管理模块
│   ├── image.go        # 镜像拉取、解析、解压
│   └── manager.go      # 镜像校验与根文件系统处理
├── network/            # 网络模块
│   ├── bridge.go       # 桥接网络
│   └── network.go      # 容器网络配置
├── storage/            # 存储模块
│   ├── metadata.go     # 元数据存储
│   └── driver.go       # 文件系统存储驱动
└── isolation/          # 隔离模块
    ├── namespace.go    # namespace
    └── filesystem.go   # 文件系统隔离
```
  
