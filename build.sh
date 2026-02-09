#!/bin/bash

# 默认值
PLATFORMS="windows/amd64 linux/amd64"
OUTPUT_DIR="."
APP_NAME="WM"
USE_UPX=true

# 默认编译所有平台
PLATFORMS="windows/amd64 linux/amd64"

# 检查 UPX
if ! command -v upx &> /dev/null; then
    echo "未找到 UPX。跳过压缩。"
    USE_UPX=false
fi

echo "下载依赖..."
go mod tidy

for PLATFORM in $PLATFORMS; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_FILENAME="wm.exe"
    else
        OUTPUT_FILENAME="wm"
    fi

    echo "正在构建 $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -o $OUTPUT_DIR/$OUTPUT_FILENAME cmd/server/main.go

    if [ $? -ne 0 ]; then
        echo "发生错误！终止脚本执行..."
        exit 1
    fi

    if [ "$USE_UPX" = true ]; then
        echo "正在使用 UPX 压缩..."
        upx $OUTPUT_DIR/$OUTPUT_FILENAME
    fi
done

echo "构建完成！构建产物在项目根目录"
