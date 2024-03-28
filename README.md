# Simple-Systemd

最基本的应用启动控制, 依赖关系空间, 解决容器中无守护进程在子进程异常退出后无法重新拉起问题

# Guide

## Install

```bash
# 在根目录下生成 build 目录, 并在其下生成各二进制
make build
```

## Usage

可通过`-h`命令查看使用说明

```bash
./build/simple-systemd -h
Usage of ./build/simple-systemd:
  -c string
        set service config file scan dir (default "./config")
```

本程序将递归的读取指定目录下所有以`.service.yaml`为后缀的文件, 将其作为服务加载到程序中，每个服务的名称为其唯一标识，必须全局唯一，配置文件格式示例如[此文件](./config/example.service.yaml)所示

# Developer

## Compile

```bash
make build
```
