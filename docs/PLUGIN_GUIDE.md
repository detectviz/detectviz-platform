# Detectviz 插件開發指南

本指南將引導您完成為 Detectviz 平台創建、配置和集成新插件的完整流程。

## 插件架構概覽

Detectviz 平台的核心設計理念是 **「一切皆插件」**。這種架構使得平台具有高度的可擴展性和可組合性，允許開發者通過創建新的插件來無縫地擴展平台的功能。

所有插件都必須實現 `Plugin` 接口，該接口定義了插件的生命週期和基本行為。

## 核心插件接口

### `Plugin` 接口

`Plugin` 接口是所有插件都必須實現的基礎接口。它定義了插件的生命週期方法，包括初始化、啟動和停止。

```go
// in pkg/domain/plugins/plugin.go
package plugins

import "context"

type Plugin interface {
    GetName() string
    Init(ctx context.Context, cfg map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

*   `GetName()`: 返回插件的唯一名稱。
*   `Init()`: 初始化插件，接收來自 `composition.yaml` 的配置。
*   `Start()`: 啟動插件，例如啟動後台任務。
*   `Stop()`: 停止插件，清理資源。

### `UIPagePlugin` 接口

`UIPagePlugin` 接口用於創建新的 Web UI 頁面。它繼承自 `Plugin` 接口，並添加了用於處理 HTTP 請求和渲染 HTML 內容的方法。

```go
// in pkg/domain/plugins/uipage.go
package plugins

type UIPagePlugin interface {
    Plugin
    GetRoute() string
    GetHTMLContent() string
    RegisterRoute(router interface{}, logger interface{}) error
}
```

*   `GetRoute()`: 返回 UI 頁面的 URL 路徑。
*   `GetHTMLContent()`: 返回 UI 頁面的 HTML 內容。
*   `RegisterRoute()`: 將 UI 頁面的路由註冊到主 HTTP 服務器。

## 創建一個新的 UI 插件

在本節中，我們將以 `HelloWorldUIPagePlugin` 為例，演示如何創建一個新的 UI 插件。

### 1. 實現 `UIPagePlugin` 接口

首先，您需要創建一個新的 Go package，並在其中創建一個實現 `UIPagePlugin` 接口的結構體。

```go
// in internal/adapters/plugins/web_ui/hello_world_ui_page.go
package web_ui

import (
    "context"
    "fmt"

    "github.com/labstack/echo/v4"

    "detectviz-platform/pkg/domain/plugins"
    "detectviz-platform/pkg/platform/contracts"
)

type HelloWorldUIPagePlugin struct {
    logger contracts.Logger
    config HelloWorldConfig
}

// ... (NewHelloWorldUIPagePlugin, GetName, Init, Start, Stop, GetRoute, GetHTMLContent, RegisterRoute)
```

### 2. 在 `composition.yaml` 中配置插件

接下來，您需要在 `configs/composition.yaml` 文件中為您的新插件添加一個配置塊。

```yaml
# in configs/composition.yaml
plugins:
  - type: hello_world_ui_page
    name: helloWorldUI
    config:
      route: "/ui/hello"
      title: "Hello World - Detectviz Platform"
      message: "歡迎使用 Detectviz 平台！"
```

### 3. 動態加載插件 (未來實現)

**注意：** 目前，插件是在 `cmd/api/main.go` 文件中手動註冊的。在未來的版本中，我們將實現一個動態插件加載器，該加載器將自動從 `composition.yaml` 文件中讀取插件配置並註冊插件。

## 結論

本指南介紹了為 Detectviz 平台創建和配置新插件的基本流程。通過遵循本指南，您可以輕鬆地擴展平台的功能，以滿足您的特定需求。
