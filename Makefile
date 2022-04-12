SHELL = /bin/bash
appEnv = prod
dockerRegistry 		:= hub.miniaixue.com/sirius-go/

rpcIntegrateShopName 		:= rpc_integrate_shop
apiIntegrateShopName 		:= api_integrate_shop
rpcUserBehaviorName 		:= rpc_user_behavior
apiUserBehaviorName 		:= api_user_behavior
appCrontabName 		 		:= app_crontab
adminCrontabName 		 	:= admin_crontab
rpcCubeName 				:= rpc_cube
apiCubeName 				:= api_cube

rpcIntegrateShopVersion 	:= v1.1.1
apiIntegrateShopVersion 	:= v1.1.1
rpcUserBehaviorVersion 		:= v1.1.1
apiUserBehaviorVersion 		:= v1.1.1
appCrontabVersion 			:= v1.1.1
adminCrontabVersion 		:= v1.1.1
rpcCubeVersion 				:= v1.1.1
apiCubeVersion 				:= v1.1.1

version := $(shell git rev-parse HEAD | cut -c 1-8)
ifeq (${appEnv}, test)
	rpcIntegrateShopVersion := ${version}
	apiIntegrateShopVersion := ${version}
	rpcUserBehaviorVersion := ${version}
	apiUserBehaviorVersion := ${version}
	appCrontabVersion := ${version}
	adminCrontabVersion := ${version}
	rpcCubeVersion := ${version}
	apiCubeVersion := ${version}
endif

rpcIntegrateShopImage	:= ${dockerRegistry}${appEnv}/${rpcIntegrateShopName}:${rpcIntegrateShopVersion}
apiIntegrateShopImage	:= ${dockerRegistry}${appEnv}/${apiIntegrateShopName}:${apiIntegrateShopVersion}
rpcUserBehaviorImage	:= ${dockerRegistry}${appEnv}/${rpcUserBehaviorName}:${rpcUserBehaviorVersion}
apiUserBehaviorImage	:= ${dockerRegistry}${appEnv}/${apiUserBehaviorName}:${apiUserBehaviorVersion}
appCrontabImage			:= ${dockerRegistry}${appEnv}/${appCrontabName}:${appCrontabVersion}
adminCrontabImage		:= ${dockerRegistry}${appEnv}/${adminCrontabName}:${adminCrontabVersion}
rpcCubeImage			:= ${dockerRegistry}${appEnv}/${rpcCubeName}:${rpcCubeVersion}
apiCubeImage			:= ${dockerRegistry}${appEnv}/${apiCubeName}:${apiCubeVersion}

all = integrate-shop app-cron user-behavior cube

local-build:
	go build -ldflags="-s -w" -o ./crontab/app/${appCrontabName} crontab/app/main.go
	go build -ldflags="-s -w" -o ./service/integrateshop/rpc/${rpcIntegrateShopName} ./service/integrateshop/rpc/integrateshop.go
	go build -ldflags="-s -w" -o ./service/integrateshop/api/${apiIntegrateShopName} ./service/integrateshop/api/server.go
	go build -ldflags="-s -w" -o ./service/userbehavior/rpc/${rpcUserBehaviorName} ./service/userbehavior/rpc/rpcuserbehavior.go
	go build -ldflags="-s -w" -o ./service/userbehavior/api/${apiUserBehaviorName} ./service/userbehavior/api/apiuserbehavior.go
	go build -ldflags="-s -w" -o ./service/cube/rpc/${rpcCubeName} ./service/cube/rpc/rpcube.go
	go build -ldflags="-s -w" -o ./service/cube/api/${apiCubeName} ./service/cube/api/cubeapi.go

#构建所有服务
docker-build: $(all) docker-build-compose docker-build-push

#积分商城
integrate-shop:
	docker build -t ${apiIntegrateShopImage} -f ./service/integrateshop/api/Dockerfile .
	docker build -t ${rpcIntegrateShopImage} -f ./service/integrateshop/rpc/Dockerfile .

#用户行为
user-behavior:
	docker build -t ${apiUserBehaviorImage} -f ./service/userbehavior/api/Dockerfile .
	docker build -t ${rpcUserBehaviorImage} -f ./service/userbehavior/rpc/Dockerfile .

# cube
cube:
	docker build -t ${apiCubeImage} -f ./service/cube/api/Dockerfile .
	docker build -t ${rpcCubeImage} -f ./service/cube/rpc/Dockerfile .
#定时任务
app-cron:
	docker build -t ${appCrontabImage} -f ./crontab/app/Dockerfile .
	docker build -t ${adminCrontabImage} -f ./crontab/admin/Dockerfile .

#推送镜像到仓库
docker-build-push:
	docker push ${rpcIntegrateShopImage}
	docker push ${apiIntegrateShopImage}
	docker push ${rpcUserBehaviorImage}
	docker push ${apiUserBehaviorImage}
	docker push ${appCrontabImage}
	docker push ${adminCrontabImage}
	docker push ${apiCubeImage}
	docker push ${rpcCubeImage}


#通过模板生成带版本compos文件
docker-build-compose:
	echo "appEnv=${appEnv}" > build.tmp
	echo "rpcIntegrateShopImage=${rpcIntegrateShopImage}" >> build.tmp
	echo "apiIntegrateShopImage=${apiIntegrateShopImage}" >> build.tmp
	echo "rpcUserBehaviorImage=${rpcUserBehaviorImage}" >> build.tmp
	echo "apiUserBehaviorImage=${apiUserBehaviorImage}" >> build.tmp
	echo "appCrontabImage=${appCrontabImage}" >> build.tmp
	echo "adminCrontabImage=${adminCrontabImage}" >> build.tmp
	echo "rpcCubeImage=${rpcCubeImage}" >> build.tmp
	echo "apiCubeImage=${apiCubeImage}" >> build.tmp
	chmod 777 build.tmp
	chmod 744 build-compose.sh && ./build-compose.sh


#清理
clean:
	-docker images|grep -E '<none>|sirius-go'|awk "{print \$$3}"|xargs docker rmi -f
	rm -f build.tmp


#单元测试
benchmark:
	cd ./test/benchmark
	go test -bench=. -benchmem -memprofile memprofile.out -cpuprofile profile.out
	#go tool pprof profile.out