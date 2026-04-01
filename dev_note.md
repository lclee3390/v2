Miniflux 開發環境筆記（Windows 11 + WSL Ubuntu 24.04 + Docker Engine）
====================================================================

這份筆記以「Docker 安裝在 WSL（非 Docker Desktop）」為前提，說明如何在此專案建立本機開發環境。

先備條件
--------

- Windows 11 已啟用 WSL2，並安裝 Ubuntu 24.04
- WSL 內已安裝 Docker Engine，且可在 WSL 直接執行 `docker`
- Go >= 1.24
- Git、make

可用下列指令快速檢查版本（在 WSL 內執行）：

```bash
go version
docker --version
git --version
make --version
```

專案路徑
--------

目前專案位於 Windows 磁碟 E:，在 WSL 的路徑是：

```
/mnt/e/02.workspace/miniflux/v2
```

如果覺得跨檔案系統效能較慢，可考慮在 WSL 家目錄重新 clone 一份，但不是必要。

步驟 1：啟動 PostgreSQL（Docker）
-------------------------------

在 WSL 開一個終端機，啟動 PostgreSQL：

```bash
docker run --rm --name miniflux2-db -p 5432:5432 \
  -e POSTGRES_DB=miniflux2 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  postgres
```

這個容器會佔用本機 5432 連接埠，並在停止後自動刪除。

步驟 2：設定環境變數
--------------------

再開一個 WSL 終端機，進入專案目錄並設定資料庫連線字串：

```bash
cd /mnt/e/02.workspace/miniflux/v2
export DATABASE_URL=postgres://postgres:postgres@localhost/miniflux2?sslmode=disable
```

步驟 3：啟動開發模式
--------------------

專案提供 `make run`，會自動執行 migration 並建立預設管理員：

```bash
make run
```

預設管理員帳密：

- 使用者名稱：`admin`
- 密碼：`test123`

啟動後請在 Windows 瀏覽器開啟：

```
http://localhost:8080
```

常用開發指令
------------

```bash
# 編譯
make miniflux

# 單元測試
make test

# Lint（需要 golangci-lint / staticcheck）
make lint

# 整合測試（需要 psql 與 nc）
make integration-test
make clean-integration-test

# 建立「lc」版 Docker 映像（tag: lc，並加上 lc 版 label）
make docker-image-lc
```

整合測試使用 `psql` 與 `nc`（netcat），若缺少可安裝：

```bash
sudo apt update
sudo apt install -y postgresql-client netcat-openbsd
```

停止服務
--------

- `make run` 以 Ctrl+C 停止
- PostgreSQL 容器以 Ctrl+C 或 `docker stop miniflux2-db` 停止

備註
----

- `make run` 已包含 `RUN_MIGRATIONS=1` 與 `CREATE_ADMIN=1` 等環境變數（見 `Makefile`）。
- 若要自行設定其他環境變數（例如 `LISTEN_ADDR`、`BASE_URL`），可在執行前先 `export`。

使用 .devcontainer 快速啟動
-------------------------

專案提供 VS Code Dev Containers 設定，會自動啟動：

- app（Go 開發容器）
- db（PostgreSQL）
- apprise（整合服務）

方式 A：VS Code Dev Containers（推薦）

1. Windows 安裝 VS Code 與 Dev Containers 擴充套件。
2. 在 WSL 確認 Docker Engine 可用。
3. 用 VS Code 開啟專案目錄。
4. 命令面板執行：`Dev Containers: Reopen in Container`。
5. 進入容器後在終端機執行：

```bash
make run
```

方式 B：不用 VS Code，直接用 docker compose

```bash
cd /mnt/e/02.workspace/miniflux/v2/.devcontainer
docker compose up -d

docker compose exec app bash
cd /workspace
make run
```

補充

- devcontainer 會將專案掛載到容器內的 `/workspace`。
- 預設管理員帳密為 `admin / test123`。
- `app` 服務使用 `sleep infinity`，所以需要手動 `make run` 才會啟動服務。

容器內 listen 位址與外部訪問
---------------------------

如果在容器內執行 `make run` 後看到：

```
Starting HTTP server listen_address=127.0.0.1:8080
```

代表只綁定容器內部的 loopback，外部即使有 port mapping 也無法訪問。
解法是讓服務綁到 `0.0.0.0`：

```bash
cd /workspace
LISTEN_ADDR=0.0.0.0:8088 make run
```

之後可用 `http://localhost:8080` 或 `http://<wsl-ip>:8080` 訪問。

若想永久化，可在 `.devcontainer/docker-compose.yml` 的 `app` 服務加上：

```yaml
environment:
  - LISTEN_ADDR=0.0.0.0:8080
```
