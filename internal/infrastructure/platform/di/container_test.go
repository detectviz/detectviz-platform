package di

import (
	"errors"
	"testing"
)

// 測試用接口和實現
type TestService interface {
	GetName() string
}

type TestServiceImpl struct {
	name string
}

func (t *TestServiceImpl) GetName() string {
	return t.name
}

type TestDependentService interface {
	GetDependencyName() string
}

type TestDependentServiceImpl struct {
	dependency TestService
}

func (t *TestDependentServiceImpl) GetDependencyName() string {
	return t.dependency.GetName()
}

func TestContainer_RegisterSingleton(t *testing.T) {
	container := NewContainer()

	// 測試註冊單例服務
	err := container.RegisterSingleton((*TestService)(nil), func() TestService {
		return &TestServiceImpl{name: "test"}
	})

	if err != nil {
		t.Errorf("RegisterSingleton failed: %v", err)
	}

	// 檢查服務是否已註冊
	if !container.IsRegistered((*TestService)(nil)) {
		t.Error("Service should be registered")
	}
}

func TestContainer_RegisterTransient(t *testing.T) {
	container := NewContainer()

	// 測試註冊瞬態服務
	err := container.RegisterTransient((*TestService)(nil), func() TestService {
		return &TestServiceImpl{name: "transient"}
	})

	if err != nil {
		t.Errorf("RegisterTransient failed: %v", err)
	}

	// 檢查服務是否已註冊
	if !container.IsRegistered((*TestService)(nil)) {
		t.Error("Service should be registered")
	}
}

func TestContainer_RegisterInstance(t *testing.T) {
	container := NewContainer()
	instance := &TestServiceImpl{name: "instance"}

	// 測試註冊實例
	err := container.RegisterInstance((*TestService)(nil), instance)

	if err != nil {
		t.Errorf("RegisterInstance failed: %v", err)
	}

	// 解析服務
	resolved, err := container.Resolve((*TestService)(nil))
	if err != nil {
		t.Errorf("Resolve failed: %v", err)
	}

	// 檢查是否是同一個實例
	if resolved != instance {
		t.Error("Resolved instance should be the same as registered instance")
	}
}

func TestContainer_ResolveSingleton(t *testing.T) {
	container := NewContainer()

	// 註冊單例服務
	err := container.RegisterSingleton((*TestService)(nil), func() TestService {
		return &TestServiceImpl{name: "singleton"}
	})
	if err != nil {
		t.Errorf("RegisterSingleton failed: %v", err)
	}

	// 第一次解析
	service1, err := container.Resolve((*TestService)(nil))
	if err != nil {
		t.Errorf("First resolve failed: %v", err)
	}

	// 第二次解析
	service2, err := container.Resolve((*TestService)(nil))
	if err != nil {
		t.Errorf("Second resolve failed: %v", err)
	}

	// 檢查是否是同一個實例
	if service1 != service2 {
		t.Error("Singleton services should be the same instance")
	}

	// 檢查服務功能
	testService := service1.(TestService)
	if testService.GetName() != "singleton" {
		t.Errorf("Expected name 'singleton', got '%s'", testService.GetName())
	}
}

func TestContainer_ResolveTransient(t *testing.T) {
	container := NewContainer()

	// 註冊瞬態服務
	err := container.RegisterTransient((*TestService)(nil), func() TestService {
		return &TestServiceImpl{name: "transient"}
	})
	if err != nil {
		t.Errorf("RegisterTransient failed: %v", err)
	}

	// 第一次解析
	service1, err := container.Resolve((*TestService)(nil))
	if err != nil {
		t.Errorf("First resolve failed: %v", err)
	}

	// 第二次解析
	service2, err := container.Resolve((*TestService)(nil))
	if err != nil {
		t.Errorf("Second resolve failed: %v", err)
	}

	// 檢查是否是不同的實例
	if service1 == service2 {
		t.Error("Transient services should be different instances")
	}
}

func TestContainer_ResolveDependency(t *testing.T) {
	container := NewContainer()

	// 註冊依賴服務
	err := container.RegisterSingleton((*TestService)(nil), func() TestService {
		return &TestServiceImpl{name: "dependency"}
	})
	if err != nil {
		t.Errorf("RegisterSingleton for dependency failed: %v", err)
	}

	// 註冊依賴於其他服務的服務
	err = container.RegisterSingleton((*TestDependentService)(nil), func(dep TestService) TestDependentService {
		return &TestDependentServiceImpl{dependency: dep}
	})
	if err != nil {
		t.Errorf("RegisterSingleton for dependent service failed: %v", err)
	}

	// 解析依賴服務
	dependentService, err := container.Resolve((*TestDependentService)(nil))
	if err != nil {
		t.Errorf("Resolve dependent service failed: %v", err)
	}

	// 檢查依賴注入是否正確
	testDependentService := dependentService.(TestDependentService)
	if testDependentService.GetDependencyName() != "dependency" {
		t.Errorf("Expected dependency name 'dependency', got '%s'", testDependentService.GetDependencyName())
	}
}

func TestContainer_ResolveWithError(t *testing.T) {
	container := NewContainer()

	// 註冊會返回錯誤的服務
	err := container.RegisterSingleton((*TestService)(nil), func() (TestService, error) {
		return nil, errors.New("factory error")
	})
	if err != nil {
		t.Errorf("RegisterSingleton failed: %v", err)
	}

	// 嘗試解析服務
	_, err = container.Resolve((*TestService)(nil))
	if err == nil {
		t.Error("Expected error from factory, but got nil")
	}

	if err.Error() != "factory error" {
		t.Errorf("Expected 'factory error', got '%s'", err.Error())
	}
}

func TestContainer_ResolveUnregistered(t *testing.T) {
	container := NewContainer()

	// 嘗試解析未註冊的服務
	_, err := container.Resolve((*TestService)(nil))
	if err == nil {
		t.Error("Expected error for unregistered service, but got nil")
	}
}

func TestContainer_GetRegisteredServices(t *testing.T) {
	container := NewContainer()

	// 註冊一些服務
	err := container.RegisterSingleton((*TestService)(nil), func() TestService {
		return &TestServiceImpl{name: "test"}
	})
	if err != nil {
		t.Errorf("RegisterSingleton failed: %v", err)
	}

	err = container.RegisterSingleton((*TestDependentService)(nil), func() TestDependentService {
		return &TestDependentServiceImpl{}
	})
	if err != nil {
		t.Errorf("RegisterSingleton failed: %v", err)
	}

	// 獲取已註冊的服務
	services := container.GetRegisteredServices()

	if len(services) != 2 {
		t.Errorf("Expected 2 registered services, got %d", len(services))
	}
}

func TestContainer_Clear(t *testing.T) {
	container := NewContainer()

	// 註冊一個服務
	err := container.RegisterSingleton((*TestService)(nil), func() TestService {
		return &TestServiceImpl{name: "test"}
	})
	if err != nil {
		t.Errorf("RegisterSingleton failed: %v", err)
	}

	// 檢查服務已註冊
	if !container.IsRegistered((*TestService)(nil)) {
		t.Error("Service should be registered before clear")
	}

	// 清空容器
	container.Clear()

	// 檢查服務已被清除
	if container.IsRegistered((*TestService)(nil)) {
		t.Error("Service should not be registered after clear")
	}

	// 檢查註冊的服務數量
	services := container.GetRegisteredServices()
	if len(services) != 0 {
		t.Errorf("Expected 0 registered services after clear, got %d", len(services))
	}
}

func TestContainer_InvalidFactory(t *testing.T) {
	container := NewContainer()

	// 嘗試註冊非函數類型的工廠
	err := container.RegisterSingleton((*TestService)(nil), "not a function")
	if err == nil {
		t.Error("Expected error for non-function factory, but got nil")
	}

	// 嘗試註冊返回值不匹配的工廠
	err = container.RegisterSingleton((*TestService)(nil), func() string {
		return "wrong type"
	})
	if err == nil {
		t.Error("Expected error for wrong return type, but got nil")
	}
}

func TestContainer_GetName(t *testing.T) {
	container := NewContainer()

	expectedName := "dependency_injection_container"
	if container.GetName() != expectedName {
		t.Errorf("Expected name '%s', got '%s'", expectedName, container.GetName())
	}
}
