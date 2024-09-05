#基础镜像
FROM hub.atomgit.com/amd64/ubuntu:23.10
#把编译后的打包进来这个镜像，放到工作目录 /app
COPY webook /app/webook
WORKDIR /app
#CMD执行命令
ENTRYPOINT ["/app/webook"]
