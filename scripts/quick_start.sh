#!/bin/bash
#Install Latest Stable IDB Release

osCheck=`uname -a`
if [[ $osCheck =~ 'x86_64' ]];then
    architecture="amd64"
elif [[ $osCheck =~ 'arm64' ]] || [[ $osCheck =~ 'aarch64' ]];then
    architecture="arm64"
# elif [[ $osCheck =~ 'armv7l' ]];then
#     architecture="armv7"
# elif [[ $osCheck =~ 'ppc64le' ]];then
#     architecture="ppc64le"
# elif [[ $osCheck =~ 's390x' ]];then
#     architecture="s390x"
else
    echo "暂不支持的系统架构，请参阅官方文档，选择受支持的系统。"
    exit 1
fi

# if [[ ! ${INSTALL_MODE} ]];then
#   INSTALL_MODE="stable"
# else
#     if [[ ${INSTALL_MODE} != "dev" && ${INSTALL_MODE} != "stable" ]];then
#         echo "请输入正确的安装模式（dev or stable）"
#         exit 1
#     fi
# fi

VERSION=$(curl -s https://static.sensdata.com/idb/release/latest)
HASH_FILE_URL="https://static.sensdata.com/idb/release/${VERSION}/checksums.txt"

if [[ "x${VERSION}" == "x" ]];then
    echo "获取最新版本失败，请稍候重试"
    exit 1
fi

package_file_name="idb-${VERSION}-linux-${architecture}.tar.gz"
package_download_url="https://static.sensdata.com/idb/release/${VERSION}/${package_file_name}"
expected_hash=$(curl -s "$HASH_FILE_URL" | grep "$package_file_name" | awk '{print $1}')

if [ -f ${package_file_name} ];then
    actual_hash=$(sha256sum "$package_file_name" | awk '{print $1}')
    if [[ "$expected_hash" == "$actual_hash" ]];then
        echo "安装包已存在，跳过下载"
        rm -rf idb-${VERSION}-linux-${architecture}
        tar zxvf ${package_file_name}
        cd idb-${VERSION}-linux-${architecture}
        /bin/bash install.sh
        exit 0
    else
        echo "已存在安装包，但是哈希值不一致，开始重新下载"
        rm -f ${package_file_name}
    fi
fi

echo "开始下载 idb ${VERSION} 版本在线安装包"
echo "安装包下载地址： ${package_download_url}"

curl -LOk -o ${package_file_name} ${package_download_url}
curl -sfL https://static.sensdata.com/idb/installation-log.sh | sh -s 1p install ${VERSION}
if [ ! -f ${package_file_name} ];then
    echo "下载安装包失败，请稍候重试。"
    exit 1
fi

tar zxvf ${package_file_name}
if [ $? != 0 ];then
    echo "下载安装包失败，请稍候重试。"
    rm -f ${package_file_name}
    exit 1
fi
cd idb-${VERSION}-linux-${architecture}

/bin/bash install.sh
