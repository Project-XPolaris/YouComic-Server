rm -rf ./pack-output
mkdir ./pack-output
mkdir ./pack-output/conf
cp -r ./conf/setup.json ./pack-output/conf/setup.json
cp -r ./assets ./pack-output/assets
cp ./main ./pack-output/youcomic
cp ./install.sh ./pack-output/install.sh
cp ./uninstall.sh ./pack-output/uninstall.sh
cp ./uninstall-service.sh ./pack-output/uninstall-service.sh
cp ./install-service.sh ./pack-output/install-service.sh
cp ./youplus.json ./pack-output/youplus.json
cp ./ulist.json ./pack-output/ulist.json
