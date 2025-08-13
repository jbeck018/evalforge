import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Settings,
  Plus,
  Save,
  Trash2,
  Edit,
  Check,
  X,
  Key,
  AlertCircle,
  TestTube,
  DollarSign,
  Zap,
  Shield,
  Copy,
  Eye,
  EyeOff
} from 'lucide-react';
import { api } from '../utils/api';

interface LLMProvider {
  id: string;
  name: string;
  provider: 'openai' | 'anthropic' | 'google' | 'azure' | 'huggingface' | 'custom';
  api_key: string;
  api_base?: string;
  models: LLMModel[];
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

interface LLMModel {
  id: string;
  name: string;
  display_name: string;
  provider_id: string;
  max_tokens: number;
  cost_per_1k_input: number;
  cost_per_1k_output: number;
  supports_functions: boolean;
  supports_vision: boolean;
  enabled: boolean;
  rate_limit?: number;
  timeout?: number;
}

interface TestResult {
  success: boolean;
  latency: number;
  message: string;
  tokens_used?: number;
  cost?: number;
}

const PROVIDER_TEMPLATES = {
  openai: {
    name: 'OpenAI',
    api_base: 'https://api.openai.com/v1',
    models: [
      { name: 'gpt-4-turbo', display_name: 'GPT-4 Turbo', max_tokens: 128000, cost_per_1k_input: 0.01, cost_per_1k_output: 0.03 },
      { name: 'gpt-4', display_name: 'GPT-4', max_tokens: 8192, cost_per_1k_input: 0.03, cost_per_1k_output: 0.06 },
      { name: 'gpt-3.5-turbo', display_name: 'GPT-3.5 Turbo', max_tokens: 16385, cost_per_1k_input: 0.001, cost_per_1k_output: 0.002 },
    ]
  },
  anthropic: {
    name: 'Anthropic',
    api_base: 'https://api.anthropic.com/v1',
    models: [
      { name: 'claude-3-opus', display_name: 'Claude 3 Opus', max_tokens: 200000, cost_per_1k_input: 0.015, cost_per_1k_output: 0.075 },
      { name: 'claude-3-sonnet', display_name: 'Claude 3 Sonnet', max_tokens: 200000, cost_per_1k_input: 0.003, cost_per_1k_output: 0.015 },
      { name: 'claude-3-haiku', display_name: 'Claude 3 Haiku', max_tokens: 200000, cost_per_1k_input: 0.00025, cost_per_1k_output: 0.00125 },
    ]
  },
  google: {
    name: 'Google AI',
    api_base: 'https://generativelanguage.googleapis.com/v1',
    models: [
      { name: 'gemini-pro', display_name: 'Gemini Pro', max_tokens: 32768, cost_per_1k_input: 0.00025, cost_per_1k_output: 0.0005 },
      { name: 'gemini-pro-vision', display_name: 'Gemini Pro Vision', max_tokens: 32768, cost_per_1k_input: 0.00025, cost_per_1k_output: 0.0005 },
    ]
  },
  azure: {
    name: 'Azure OpenAI',
    api_base: '', // User must provide
    models: [
      { name: 'gpt-4', display_name: 'GPT-4 (Azure)', max_tokens: 8192, cost_per_1k_input: 0.03, cost_per_1k_output: 0.06 },
      { name: 'gpt-35-turbo', display_name: 'GPT-3.5 Turbo (Azure)', max_tokens: 16385, cost_per_1k_input: 0.001, cost_per_1k_output: 0.002 },
    ]
  },
  huggingface: {
    name: 'Hugging Face',
    api_base: 'https://api-inference.huggingface.co',
    models: [
      { name: 'meta-llama/Llama-2-70b-chat-hf', display_name: 'Llama 2 70B', max_tokens: 4096, cost_per_1k_input: 0.0007, cost_per_1k_output: 0.0007 },
      { name: 'mistralai/Mixtral-8x7B-Instruct-v0.1', display_name: 'Mixtral 8x7B', max_tokens: 32768, cost_per_1k_input: 0.0007, cost_per_1k_output: 0.0007 },
    ]
  },
  custom: {
    name: 'Custom Provider',
    api_base: '',
    models: []
  }
};

const LLMConfigPage: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [selectedProvider, setSelectedProvider] = useState<LLMProvider | null>(null);
  const [isAddingProvider, setIsAddingProvider] = useState(false);
  const [newProvider, setNewProvider] = useState<Partial<LLMProvider>>({
    provider: 'openai',
    enabled: true
  });
  const [showApiKey, setShowApiKey] = useState<string | null>(null);
  const [testingProvider, setTestingProvider] = useState<string | null>(null);
  const [testResults, setTestResults] = useState<Record<string, TestResult>>({});

  // Fetch LLM providers
  const { data: providers = [], isLoading } = useQuery({
    queryKey: ['llm-providers'],
    queryFn: async () => {
      try {
        const response = await api.get('/api/llm/providers');
        return response.data.providers || [];
      } catch {
        // Return mock data if endpoint doesn't exist yet
        return [];
      }
    }
  });

  // Create provider mutation
  const createProviderMutation = useMutation({
    mutationFn: async (provider: Partial<LLMProvider>) => {
      return api.post('/api/llm/providers', provider);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm-providers'] });
      setIsAddingProvider(false);
      setNewProvider({ provider: 'openai', enabled: true });
    }
  });

  // Update provider mutation
  const updateProviderMutation = useMutation({
    mutationFn: async ({ id, ...data }: Partial<LLMProvider>) => {
      return api.put(`/api/llm/providers/${id}`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm-providers'] });
      setSelectedProvider(null);
    }
  });

  // Delete provider mutation
  const deleteProviderMutation = useMutation({
    mutationFn: async (id: string) => {
      return api.delete(`/api/llm/providers/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm-providers'] });
      setSelectedProvider(null);
    }
  });

  // Test provider connection
  const testProvider = async (provider: LLMProvider | Partial<LLMProvider>) => {
    const id = provider.id || 'new';
    setTestingProvider(id);
    
    try {
      const response = await api.post('/api/llm/test', {
        provider: provider.provider,
        api_key: provider.api_key,
        api_base: provider.api_base
      });
      
      setTestResults({
        ...testResults,
        [id]: response.data
      });
    } catch (error) {
      setTestResults({
        ...testResults,
        [id]: {
          success: false,
          latency: 0,
          message: 'Connection failed'
        }
      });
    } finally {
      setTestingProvider(null);
    }
  };

  const handleAddProvider = () => {
    if (!newProvider.name || !newProvider.api_key) return;
    
    const template = PROVIDER_TEMPLATES[newProvider.provider as keyof typeof PROVIDER_TEMPLATES];
    const providerData = {
      ...newProvider,
      name: newProvider.name || template.name,
      api_base: newProvider.api_base || template.api_base,
      models: template.models.map(model => ({
        ...model,
        id: `${newProvider.provider}-${model.name}`,
        provider_id: newProvider.id,
        enabled: true,
        supports_functions: ['gpt-4', 'gpt-3.5-turbo'].includes(model.name),
        supports_vision: model.name.includes('vision') || model.name === 'gpt-4-turbo'
      }))
    };
    
    // For now, just save to local state
    // In production, this would call the API
    createProviderMutation.mutate(providerData);
  };

  const handleUpdateModel = (providerId: string, modelId: string, updates: Partial<LLMModel>) => {
    const provider = providers.find(p => p.id === providerId);
    if (!provider) return;
    
    const updatedModels = provider.models.map(model => 
      model.id === modelId ? { ...model, ...updates } : model
    );
    
    updateProviderMutation.mutate({
      id: providerId,
      models: updatedModels
    });
  };

  const copyApiKey = (key: string) => {
    navigator.clipboard.writeText(key);
    // Show toast notification
  };

  const maskApiKey = (key: string) => {
    if (!key) return '';
    return key.substring(0, 7) + '...' + key.substring(key.length - 4);
  };

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 flex items-center">
              <Settings className="h-6 w-6 mr-2 text-indigo-600" />
              LLM Provider Configuration
            </h1>
            <p className="mt-1 text-sm text-gray-500">
              Configure and manage your LLM providers and models
            </p>
          </div>
          <button
            onClick={() => setIsAddingProvider(true)}
            className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700 flex items-center"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Provider
          </button>
        </div>
      </div>

      {/* Add Provider Form */}
      {isAddingProvider && (
        <div className="bg-white shadow rounded-lg p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Add New Provider</h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Provider Type
              </label>
              <select
                value={newProvider.provider}
                onChange={(e) => setNewProvider({ ...newProvider, provider: e.target.value as any })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              >
                {Object.entries(PROVIDER_TEMPLATES).map(([key, template]) => (
                  <option key={key} value={key}>{template.name}</option>
                ))}
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Display Name
              </label>
              <input
                type="text"
                value={newProvider.name || ''}
                onChange={(e) => setNewProvider({ ...newProvider, name: e.target.value })}
                placeholder={PROVIDER_TEMPLATES[newProvider.provider as keyof typeof PROVIDER_TEMPLATES]?.name}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                API Key
              </label>
              <input
                type="password"
                value={newProvider.api_key || ''}
                onChange={(e) => setNewProvider({ ...newProvider, api_key: e.target.value })}
                placeholder="sk-..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                API Base URL (optional)
              </label>
              <input
                type="text"
                value={newProvider.api_base || ''}
                onChange={(e) => setNewProvider({ ...newProvider, api_base: e.target.value })}
                placeholder={PROVIDER_TEMPLATES[newProvider.provider as keyof typeof PROVIDER_TEMPLATES]?.api_base}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>
          </div>
          
          <div className="mt-4 flex justify-end space-x-3">
            <button
              onClick={() => {
                setIsAddingProvider(false);
                setNewProvider({ provider: 'openai', enabled: true });
              }}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              onClick={() => testProvider(newProvider)}
              disabled={!newProvider.api_key || testingProvider === 'new'}
              className="px-4 py-2 border border-indigo-600 rounded-md text-sm font-medium text-indigo-600 hover:bg-indigo-50 flex items-center"
            >
              <TestTube className="h-4 w-4 mr-2" />
              {testingProvider === 'new' ? 'Testing...' : 'Test Connection'}
            </button>
            <button
              onClick={handleAddProvider}
              disabled={!newProvider.name || !newProvider.api_key}
              className="px-4 py-2 bg-indigo-600 text-white rounded-md text-sm font-medium hover:bg-indigo-700 flex items-center"
            >
              <Save className="h-4 w-4 mr-2" />
              Save Provider
            </button>
          </div>
          
          {testResults['new'] && (
            <div className={`mt-4 p-4 rounded-md ${testResults['new'].success ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center">
                {testResults['new'].success ? (
                  <Check className="h-5 w-5 text-green-600 mr-2" />
                ) : (
                  <X className="h-5 w-5 text-red-600 mr-2" />
                )}
                <p className={`text-sm ${testResults['new'].success ? 'text-green-800' : 'text-red-800'}`}>
                  {testResults['new'].message}
                  {testResults['new'].latency > 0 && ` (${testResults['new'].latency}ms)`}
                </p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Providers List */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {providers.length === 0 && !isAddingProvider ? (
          <div className="col-span-2 bg-white shadow rounded-lg p-12 text-center">
            <Settings className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No providers configured</h3>
            <p className="text-gray-500 mb-4">Add your first LLM provider to start monitoring and evaluating</p>
            <button
              onClick={() => setIsAddingProvider(true)}
              className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700 inline-flex items-center"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add Provider
            </button>
          </div>
        ) : (
          providers.map((provider) => (
            <div key={provider.id} className="bg-white shadow rounded-lg">
              <div className="p-6 border-b border-gray-200">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <div className={`h-10 w-10 rounded-lg flex items-center justify-center mr-3 ${
                      provider.enabled ? 'bg-green-100' : 'bg-gray-100'
                    }`}>
                      <Settings className={`h-6 w-6 ${
                        provider.enabled ? 'text-green-600' : 'text-gray-400'
                      }`} />
                    </div>
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900">{provider.name}</h3>
                      <p className="text-sm text-gray-500">{provider.provider}</p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => testProvider(provider)}
                      disabled={testingProvider === provider.id}
                      className="p-2 text-gray-400 hover:text-gray-600"
                    >
                      <TestTube className="h-5 w-5" />
                    </button>
                    <button
                      onClick={() => setSelectedProvider(provider)}
                      className="p-2 text-gray-400 hover:text-gray-600"
                    >
                      <Edit className="h-5 w-5" />
                    </button>
                    <button
                      onClick={() => deleteProviderMutation.mutate(provider.id)}
                      className="p-2 text-gray-400 hover:text-red-600"
                    >
                      <Trash2 className="h-5 w-5" />
                    </button>
                  </div>
                </div>
                
                {/* API Key Display */}
                <div className="mt-4 p-3 bg-gray-50 rounded-md">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <Key className="h-4 w-4 text-gray-400 mr-2" />
                      <span className="text-sm text-gray-600">
                        {showApiKey === provider.id ? provider.api_key : maskApiKey(provider.api_key)}
                      </span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <button
                        onClick={() => setShowApiKey(showApiKey === provider.id ? null : provider.id)}
                        className="p-1 text-gray-400 hover:text-gray-600"
                      >
                        {showApiKey === provider.id ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                      <button
                        onClick={() => copyApiKey(provider.api_key)}
                        className="p-1 text-gray-400 hover:text-gray-600"
                      >
                        <Copy className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                </div>
                
                {/* Test Result */}
                {testResults[provider.id] && (
                  <div className={`mt-4 p-3 rounded-md ${
                    testResults[provider.id].success ? 'bg-green-50' : 'bg-red-50'
                  }`}>
                    <div className="flex items-center">
                      {testResults[provider.id].success ? (
                        <Check className="h-4 w-4 text-green-600 mr-2" />
                      ) : (
                        <X className="h-4 w-4 text-red-600 mr-2" />
                      )}
                      <span className={`text-sm ${
                        testResults[provider.id].success ? 'text-green-800' : 'text-red-800'
                      }`}>
                        {testResults[provider.id].message}
                      </span>
                    </div>
                  </div>
                )}
              </div>
              
              {/* Models */}
              <div className="p-6">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Available Models</h4>
                <div className="space-y-2">
                  {provider.models?.map((model) => (
                    <div key={model.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-md">
                      <div className="flex items-center">
                        <input
                          type="checkbox"
                          checked={model.enabled}
                          onChange={(e) => handleUpdateModel(provider.id, model.id, { enabled: e.target.checked })}
                          className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded mr-3"
                        />
                        <div>
                          <p className="text-sm font-medium text-gray-900">{model.display_name}</p>
                          <div className="flex items-center space-x-4 mt-1">
                            <span className="text-xs text-gray-500 flex items-center">
                              <DollarSign className="h-3 w-3 mr-1" />
                              ${model.cost_per_1k_input}/1k in, ${model.cost_per_1k_output}/1k out
                            </span>
                            <span className="text-xs text-gray-500 flex items-center">
                              <Zap className="h-3 w-3 mr-1" />
                              {model.max_tokens.toLocaleString()} tokens
                            </span>
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        {model.supports_functions && (
                          <span className="px-2 py-1 text-xs bg-blue-100 text-blue-800 rounded">Functions</span>
                        )}
                        {model.supports_vision && (
                          <span className="px-2 py-1 text-xs bg-purple-100 text-purple-800 rounded">Vision</span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Info Section */}
      <div className="mt-8 bg-blue-50 rounded-lg p-6">
        <div className="flex items-start">
          <AlertCircle className="h-5 w-5 text-blue-600 mt-0.5 mr-3" />
          <div>
            <h3 className="text-sm font-medium text-blue-800">Security Best Practices</h3>
            <ul className="mt-2 text-sm text-blue-700 space-y-1">
              <li>• API keys are encrypted at rest and in transit</li>
              <li>• Use environment-specific keys and rotate them regularly</li>
              <li>• Enable rate limiting to prevent abuse</li>
              <li>• Monitor usage and costs through the analytics dashboard</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LLMConfigPage;