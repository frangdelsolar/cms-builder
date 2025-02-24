package resourcemanager

type ResourceManager struct {
	Resources map[string]interface{}
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		Resources: make(map[string]interface{}),
	}
}

type ResourceConfig struct {
	Model    interface{}
	Handlers *ApiHandlers
}

func (r *ResourceManager) AddResource(resourceInput *ResourceConfig) {
}
