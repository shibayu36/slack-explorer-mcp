# MCP Registry公開 実装計画

## 概要
slack-explorer-mcpをMCP公式レジストリ（registry.modelcontextprotocol.io）に公開する。

## 公開方式
- **Docker/OCI方式**を採用
- 既存のghcr.ioへのDockerイメージ公開インフラを活用
- GitHub ActionsのOIDC認証でsecretsの追加不要

## 必要な変更

### Commit 1: DockerfileにMCP Registry用のLABELを追加

**ファイル:** `Dockerfile`

**変更内容:**
```dockerfile
# Runtime stageのセクションに追加
LABEL io.modelcontextprotocol.server.name="io.github.shibayu36/slack-explorer-mcp"
```

**理由:**
MCP RegistryがDocker/OCIイメージの所有権を検証するために必要。
このLABELの値がserver.jsonのnameと一致している必要がある。

---

### Commit 2: server.jsonを作成

**ファイル:** `server.json`（新規作成）

**内容:**
```json
{
  "$schema": "https://static.modelcontextprotocol.io/schemas/2025-12-11/server.schema.json",
  "name": "io.github.shibayu36/slack-explorer-mcp",
  "title": "Slack Explorer MCP",
  "description": "MCP server for searching and exploring Slack messages",
  "repository": {
    "url": "https://github.com/shibayu36/slack-explorer-mcp",
    "source": "github"
  },
  "version": "0.8.0",
  "packages": [
    {
      "registryType": "oci",
      "identifier": "ghcr.io/shibayu36/slack-explorer-mcp:0.8.0",
      "transport": {
        "type": "stdio"
      }
    }
  ]
}
```

**注意:**
- `name`はDockerfileのLABELと一致させる
- `version`と`packages[].identifier`のタグはrelease.sh実行時に自動更新される

---

### Commit 3: release.shにserver.jsonのバージョン更新を追加

**ファイル:** `release.sh`

**追加内容（main.goのバージョン更新の後に追加）:**
```bash
# Update version in server.json
echo "Updating server.json version..."
jq --arg v "$VERSION" --arg id "ghcr.io/shibayu36/slack-explorer-mcp:$VERSION" \
  '.version = $v | .packages[0].identifier = $id' server.json > server.tmp && mv server.tmp server.json
```

**git addの変更:**
```bash
git add main.go server.json
```

**理由:**
リポジトリ上のserver.jsonが常に最新バージョンを反映するようにする。

---

### Commit 4: publish-docker.ymlにMCPレジストリ公開ステップを統合

**ファイル:** `.github/workflows/publish-docker.yml`

**追加内容（Dockerイメージ公開後に追加）:**
```yaml
      # MCP Registry公開
      - name: Install mcp-publisher
        run: |
          curl -L "https://github.com/modelcontextprotocol/registry/releases/latest/download/mcp-publisher_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz" | tar xz mcp-publisher

      - name: Authenticate to MCP Registry
        run: ./mcp-publisher login github-oidc

      - name: Publish to MCP Registry
        run: ./mcp-publisher publish
```

**permissions追加:**
```yaml
permissions:
  contents: read
  packages: write
  id-token: write  # OIDC認証に必要
```

**注意:**
- server.jsonは既にrelease.shで更新されているので、CI時の動的更新は不要

---

## 実行手順

1. 上記4つのcommitを作成
2. mainブランチにpush
3. GitHub Actionsのworkflow_dispatchで手動実行
4. MCP Registryへの公開を確認:
   ```bash
   curl "https://registry.modelcontextprotocol.io/v0.1/servers?search=io.github.shibayu36/slack-explorer-mcp"
   ```

## 将来のリリースフロー

1. `./release.sh 0.9.0` を実行
   - main.goのVersion更新
   - server.jsonのversion更新
   - コミット & タグ作成 & プッシュ
2. GitHub Actionsが自動実行
   - Dockerイメージがghcr.ioに公開される
   - MCP Registryに公開される

## 参考ドキュメント
- https://github.com/modelcontextprotocol/registry/blob/main/docs/modelcontextprotocol-io/package-types.mdx
- https://github.com/modelcontextprotocol/registry/blob/main/docs/modelcontextprotocol-io/github-actions.mdx
