package di

import (
	"fmt"
	"reflect"
	"sync"
)

// ServiceScope 定義服務的生命週期範圍
type ServiceScope int

const (
	ScopeSingleton ServiceScope = iota // 單例模式
	ScopeTransient                     // 每次創建新實例
)

// ServiceDescriptor 描述一個服務的註冊信息
type ServiceDescriptor struct {
	ServiceType reflect.Type
	Factory     interface{}
	Scope       ServiceScope
	Instance    interface{}
}

// Container 是依賴注入容器
// AI_PLUGIN_TYPE: "dependency_injection_container"
// AI_IMPL_PACKAGE: "detectviz-platform/internal/infrastructure/platform/di"
// AI_IMPL_CONSTRUCTOR: "NewContainer"
type Container struct {
	services map[reflect.Type]*ServiceDescriptor
	mu       sync.RWMutex
}

// NewContainer 創建新的依賴注入容器
func NewContainer() *Container {
	return &Container{
		services: make(map[reflect.Type]*ServiceDescriptor),
	}
}

// RegisterSingleton 註冊單例服務
func (c *Container) RegisterSingleton(serviceType interface{}, factory interface{}) error {
	return c.register(serviceType, factory, ScopeSingleton)
}

// RegisterTransient 註冊瞬態服務
func (c *Container) RegisterTransient(serviceType interface{}, factory interface{}) error {
	return c.register(serviceType, factory, ScopeTransient)
}

// RegisterInstance 註冊已存在的實例作為單例
func (c *Container) RegisterInstance(serviceType interface{}, instance interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	serviceTypeReflect := reflect.TypeOf(serviceType).Elem()
	instanceType := reflect.TypeOf(instance)

	// 檢查實例是否實現了指定的接口
	if !instanceType.Implements(serviceTypeReflect) {
		return fmt.Errorf("instance of type %v does not implement interface %v", instanceType, serviceTypeReflect)
	}

	c.services[serviceTypeReflect] = &ServiceDescriptor{
		ServiceType: serviceTypeReflect,
		Factory:     nil,
		Scope:       ScopeSingleton,
		Instance:    instance,
	}

	return nil
}

// register 內部註冊方法
func (c *Container) register(serviceType interface{}, factory interface{}, scope ServiceScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	serviceTypeReflect := reflect.TypeOf(serviceType).Elem()
	factoryType := reflect.TypeOf(factory)

	// 驗證工廠函數
	if factoryType.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function")
	}

	// 檢查工廠函數的返回值
	if factoryType.NumOut() < 1 || factoryType.NumOut() > 2 {
		return fmt.Errorf("factory function must return 1 or 2 values (service, error)")
	}

	// 檢查第一個返回值是否實現了服務接口
	returnType := factoryType.Out(0)
	if !returnType.Implements(serviceTypeReflect) {
		return fmt.Errorf("factory return type %v does not implement interface %v", returnType, serviceTypeReflect)
	}

	// 如果有第二個返回值，必須是 error 類型
	if factoryType.NumOut() == 2 {
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		if !factoryType.Out(1).Implements(errorType) {
			return fmt.Errorf("second return value must be error type")
		}
	}

	c.services[serviceTypeReflect] = &ServiceDescriptor{
		ServiceType: serviceTypeReflect,
		Factory:     factory,
		Scope:       scope,
		Instance:    nil,
	}

	return nil
}

// Resolve 解析服務實例
func (c *Container) Resolve(serviceType interface{}) (interface{}, error) {
	c.mu.RLock()
	serviceTypeReflect := reflect.TypeOf(serviceType).Elem()
	descriptor, exists := c.services[serviceTypeReflect]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service of type %v is not registered", serviceTypeReflect)
	}

	// 如果是單例且已經有實例，直接返回
	if descriptor.Scope == ScopeSingleton && descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	// 如果沒有工廠函數但有實例，返回實例
	if descriptor.Factory == nil && descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	// 創建新實例
	instance, err := c.createInstance(descriptor)
	if err != nil {
		return nil, err
	}

	// 如果是單例，保存實例
	if descriptor.Scope == ScopeSingleton {
		c.mu.Lock()
		descriptor.Instance = instance
		c.mu.Unlock()
	}

	return instance, nil
}

// createInstance 創建服務實例
func (c *Container) createInstance(descriptor *ServiceDescriptor) (interface{}, error) {
	factoryValue := reflect.ValueOf(descriptor.Factory)
	factoryType := factoryValue.Type()

	// 準備參數
	args := make([]reflect.Value, factoryType.NumIn())
	for i := 0; i < factoryType.NumIn(); i++ {
		paramType := factoryType.In(i)

		// 嘗試解析參數
		paramInstance, err := c.resolveByType(paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parameter %d of type %v: %w", i, paramType, err)
		}

		args[i] = reflect.ValueOf(paramInstance)
	}

	// 調用工廠函數
	results := factoryValue.Call(args)

	// 檢查錯誤
	if len(results) == 2 && !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface(), nil
}

// resolveByType 根據類型解析服務
func (c *Container) resolveByType(serviceType reflect.Type) (interface{}, error) {
	c.mu.RLock()
	descriptor, exists := c.services[serviceType]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service of type %v is not registered", serviceType)
	}

	// 如果是單例且已經有實例，直接返回
	if descriptor.Scope == ScopeSingleton && descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	// 如果沒有工廠函數但有實例，返回實例
	if descriptor.Factory == nil && descriptor.Instance != nil {
		return descriptor.Instance, nil
	}

	// 創建新實例
	instance, err := c.createInstance(descriptor)
	if err != nil {
		return nil, err
	}

	// 如果是單例，保存實例
	if descriptor.Scope == ScopeSingleton {
		c.mu.Lock()
		descriptor.Instance = instance
		c.mu.Unlock()
	}

	return instance, nil
}

// GetRegisteredServices 獲取所有已註冊的服務類型
func (c *Container) GetRegisteredServices() []reflect.Type {
	c.mu.RLock()
	defer c.mu.RUnlock()

	services := make([]reflect.Type, 0, len(c.services))
	for serviceType := range c.services {
		services = append(services, serviceType)
	}

	return services
}

// IsRegistered 檢查服務是否已註冊
func (c *Container) IsRegistered(serviceType interface{}) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	serviceTypeReflect := reflect.TypeOf(serviceType).Elem()
	_, exists := c.services[serviceTypeReflect]
	return exists
}

// Clear 清空容器中的所有服務
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[reflect.Type]*ServiceDescriptor)
}

// GetName 返回容器名稱
func (c *Container) GetName() string {
	return "dependency_injection_container"
}
