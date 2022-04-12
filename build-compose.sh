#!/bin/bash
##生产docker-compose文件

#模板文件
tmpl=./docker-compose.tmpl.yaml
#替换后文件
target=./docker-compose.yaml
#环境prod、test
appEnv=test
#镜像地址
registry="hub.miniaixue.com/sirius-go/${appEnv}/"
#临时版本号
version=$(git rev-parse HEAD | cut -c 1-8)

#默认项目
rpcIntegrateShopImage="${registry}rpc_integrate_shop:${version}"
apiIntegrateShopImage="${registry}api_integrate_shop:${version}"
rpcUserBehaviorImage="${registry}rpc_user_behavior:${version}"
apiUserBehaviorImage="${registry}api_user_behavior:${version}"
rpcCubeImage="${registry}rpc_cube:${version}"
apiCubeImage="${registry}api_cube:${version}"
appCrontabImage="${registry}app_crontab:${version}"
adminCrontabImage="${registry}admin_crontab:${version}"

#副本数
apiIntegrateShopReplicas=3
rpcIntegrateShopReplicas=3
apiUserBehaviorReplicas=3
rpcUserBehaviorReplicas=3
apiCubeReplicas=3
rpcCubeReplicas=3
appCrontabReplicas=1
adminCrontabReplicas=1

env=.${appEnv}
#提取待替换变量${.*}
args=`cat $tmpl |grep -E '\\${\\S+}' |awk -F '{' '{print $2}'|awk -F '}' '{print $1}'|sort|uniq`


#获取外部版本配置文件
if [ -f "./build.tmp" ]; then
    for line in `cat ./build.tmp`
    do
        eval "$line"
    done
fi

#测试环境调整
if [ $appEnv == test ]; then
    apiIntegrateShopReplicas=1
    rpcIntegrateShopReplicas=1
    apiUserBehaviorReplicas=1
    rpcUserBehaviorReplicas=1
    appCrontabReplicas=1
    adminCrontabReplicas=1
    apiCubeReplicas=1
    rpcCubeReplicas=1
fi
if [ $appEnv == prod ]; then
    env=""
fi

#生产文件等待替换
cp -f $tmpl $target

#参数替换
for arg in ${args[*]}
do
    eval value='$'"${arg}"
    sed -i "s~\${$arg}~$value~g" $target
done