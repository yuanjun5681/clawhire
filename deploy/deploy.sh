#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "用法: $0 <user@host> [选项]"
  echo ""
  echo "选项:"
  echo "  -t, --tag <tag>       镜像 tag（默认: latest）"
  echo "  -d, --dir <path>      远程部署目录（默认: /opt/clawhire）"
  echo "  --setup               首次部署：传输 compose 与 .env.example"
  echo "  --compose-only        仅更新 compose 文件并重启"
  echo "  -h, --help            显示帮助"
  exit 1
}

REMOTE_HOST=""
TAG=""
REMOTE_DIR="/opt/clawhire"
SETUP=false
COMPOSE_ONLY=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    -t|--tag) TAG="$2"; shift 2 ;;
    -d|--dir) REMOTE_DIR="$2"; shift 2 ;;
    --setup) SETUP=true; shift ;;
    --compose-only) COMPOSE_ONLY=true; shift ;;
    -h|--help) usage ;;
    -*) echo "未知选项: $1"; usage ;;
    *)
      if [[ -z "$REMOTE_HOST" ]]; then
        REMOTE_HOST="$1"
        shift
      else
        echo "多余参数: $1"
        usage
      fi
      ;;
  esac
done

[[ -z "$REMOTE_HOST" ]] && usage
[[ -z "$TAG" ]] && TAG="latest"

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGES_FILE="clawhire-images-${TAG}.tar.gz"

if $SETUP; then
  echo "==> 首次部署: 初始化远程目录 ${REMOTE_DIR}"
  ssh "${REMOTE_HOST}" "mkdir -p ${REMOTE_DIR}"
  scp "${PROJECT_ROOT}/docker-compose.prod.yml" "${REMOTE_HOST}:${REMOTE_DIR}/docker-compose.yml"
  scp "${PROJECT_ROOT}/deploy/.env.example" "${REMOTE_HOST}:${REMOTE_DIR}/.env.example"
  echo ""
  echo "compose 和 .env.example 已传输。请在远程服务器配置 .env:"
  echo "  ssh ${REMOTE_HOST}"
  echo "  cd ${REMOTE_DIR}"
  echo "  cp .env.example .env"
  echo "  vim .env"
  echo ""
  echo "配置完成后运行:"
  echo "  $0 ${REMOTE_HOST} -t ${TAG}"
  exit 0
fi

if $COMPOSE_ONLY; then
  echo "==> 更新 compose 文件"
  scp "${PROJECT_ROOT}/docker-compose.prod.yml" "${REMOTE_HOST}:${REMOTE_DIR}/docker-compose.yml"
  ssh "${REMOTE_HOST}" "cd ${REMOTE_DIR} && TAG=${TAG} docker compose up -d"
  echo "==> compose 文件已更新并重启服务"
  exit 0
fi

echo "==> 构建镜像 (tag: ${TAG})"
docker build --platform linux/amd64 -t "clawhire/backend:${TAG}" "${PROJECT_ROOT}/backend"
docker build --platform linux/amd64 -t "clawhire/frontend:${TAG}" "${PROJECT_ROOT}/frontend"
docker build --platform linux/amd64 -t "clawhire/clawsynapse:${TAG}" \
  --build-arg CLAWSYNAPSE_REF="${CLAWSYNAPSE_REF:-main}" \
  --build-arg CACHE_BUST="$(date +%s)" \
  "${PROJECT_ROOT}/deploy/clawsynapse"

echo "==> 导出镜像到 /tmp/${IMAGES_FILE}"
docker save \
  "clawhire/backend:${TAG}" \
  "clawhire/frontend:${TAG}" \
  "clawhire/clawsynapse:${TAG}" \
  | gzip > "/tmp/${IMAGES_FILE}"

echo "==> 传输镜像到 ${REMOTE_HOST}:${REMOTE_DIR}/"
if ssh "${REMOTE_HOST}" "command -v rsync >/dev/null 2>&1"; then
  rsync -avz --progress --partial "/tmp/${IMAGES_FILE}" "${REMOTE_HOST}:${REMOTE_DIR}/${IMAGES_FILE}"
else
  echo "(远程无 rsync, 回退到 scp)"
  scp "/tmp/${IMAGES_FILE}" "${REMOTE_HOST}:${REMOTE_DIR}/${IMAGES_FILE}"
fi

echo "==> 更新远程 compose 文件"
scp "${PROJECT_ROOT}/docker-compose.prod.yml" "${REMOTE_HOST}:${REMOTE_DIR}/docker-compose.yml"

echo "==> 在远程服务器加载镜像并启动服务"
ssh "${REMOTE_HOST}" bash -s <<EOF
set -euo pipefail
cd ${REMOTE_DIR}

echo "--- 加载镜像"
docker load < ${IMAGES_FILE}

echo "--- 更新服务"
TAG=${TAG} docker compose up -d

echo "--- 清理镜像包"
rm -f ${IMAGES_FILE}

echo "--- 清理旧镜像"
docker image prune -f

echo "--- 服务状态"
docker compose ps
EOF

rm -f "/tmp/${IMAGES_FILE}"
echo "==> 部署完成! (tag: ${TAG})"
