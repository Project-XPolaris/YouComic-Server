# 编译

## YouComic Server 文件清单
编译后的完成目录结构为:
```
assets
conf
  |-setup.json

(二进制文件)
```
如果是手动编译，请将上面的一些文件也加入至文件夹

## Windows

由于依赖中含有`sqlite`在做跨平台编译时需要CGO的支持，需要配置CGO

Windows用户可以使用powershell运行`./build.ps1`来选择对应的平台版本。

### WSL(推荐)

使用WSL编译linux版本会简单许多，配置Go的环境后，启动`./build.sh`即可。

## Linux

执行脚本`./build.sh`即可

## 其他

其他平台可以根据自己需要配置go compile，将源码编译为二进制文件后即可使用。

## Docker
Docker 编译可以执行脚本`./build.sh`即可。
docker需要包含`./docker`文件夹下的配置文件到输出的根目录，如果是通过脚本编译，编译脚本会自动拷贝相应的文件。