import React, { createContext, useContext, useState } from 'react';

export interface ModelOption {
  label: string;
  value: string;
}

export const AVAILABLE_MODELS: ModelOption[] = [
  // Existing models
//   { label: 'Claude 3.5', value: 'claude-3.5' },
  { label: 'Claude 3.7', value: 'claude-3.7-sonnet' },
  { label: 'GPT 4', value: 'gpt-4' },
  
  // New models
//   { label: 'AI21 Jamba 1.5 Large', value: 'ai21-jamba-1.5-large' },
//   { label: 'AI21 Jamba 1.5 Mini', value: 'ai21-jamba-1.5-mini' },
//   { label: 'Codestral 25.01', value: 'codestral-2501' },
//   { label: 'Cohere Command R', value: 'cohere-command-r' },
//   { label: 'Cohere Command R 08-2024', value: 'cohere-command-r-08-2024' },
//   { label: 'Cohere Command R+', value: 'cohere-command-r-plus' },
//   { label: 'Cohere Command R+ 08-2024', value: 'cohere-command-r-plus-08-2024' },
//   { label: 'DeepSeek-R1', value: 'deepseek-r1' },
//   { label: 'JAIS 30b Chat', value: 'jais-30b-chat' },
//   { label: 'Llama-3.2-11B-Vision-Instruct', value: 'llama-3.2-11b-vision-instruct' },
//   { label: 'Llama-3.2-90B-Vision-Instruct', value: 'llama-3.2-90b-vision-instruct' },
//   { label: 'Llama-3.3-70B-Instruct', value: 'llama-3.3-70b-instruct' },
//   { label: 'Meta-Llama-3-70B-Instruct', value: 'meta-llama-3-70b-instruct' },
//   { label: 'Meta-Llama-3-8B-Instruct', value: 'meta-llama-3-8b-instruct' },
//   { label: 'Meta-Llama-3.1-405B-Instruct', value: 'meta-llama-3.1-405b-instruct' },
//   { label: 'Meta-Llama-3.1-70B-Instruct', value: 'meta-llama-3.1-70b-instruct' },
//   { label: 'Meta-Llama-3.1-8B-Instruct', value: 'meta-llama-3.1-8b-instruct' },
//   { label: 'Ministral 3B', value: 'ministral-3b' },
//   { label: 'Mistral Large', value: 'mistral-large' },
//   { label: 'Mistral Large (2407)', value: 'mistral-large-2407' },
//   { label: 'Mistral Large 24.11', value: 'mistral-large-2411' },
//   { label: 'Mistral Nemo', value: 'mistral-nemo' },
//   { label: 'Mistral Small', value: 'mistral-small' },
  { label: 'OpenAI GPT-4o', value: 'gpt-4o' },
  { label: 'OpenAI GPT-4o mini', value: 'gpt-4o-mini' },
  { label: 'OpenAI o1', value: 'o1' },
  { label: 'OpenAI o1-mini', value: 'o1-mini' },
  { label: 'OpenAI o1-preview', value: 'o1-preview' },
  { label: 'OpenAI o3-mini', value: 'o3-mini' },
//   { label: 'Phi-3-medium instruct (128k)', value: 'phi-3-medium-128k-instruct' },
//   { label: 'Phi-3-medium instruct (4k)', value: 'phi-3-medium-4k-instruct' },
//   { label: 'Phi-3-mini instruct (128k)', value: 'phi-3-mini-128k-instruct' },
//   { label: 'Phi-3-mini instruct (4k)', value: 'phi-3-mini-4k-instruct' },
//   { label: 'Phi-3-small instruct (128k)', value: 'phi-3-small-128k-instruct' },
//   { label: 'Phi-3-small instruct (8k)', value: 'phi-3-small-8k-instruct' },
//   { label: 'Phi-3.5-mini instruct (128k)', value: 'phi-3.5-mini-instruct' },
//   { label: 'Phi-3.5-MoE instruct (128k)', value: 'phi-3.5-moe-instruct' },
//   { label: 'Phi-3.5-vision instruct (128k)', value: 'phi-3.5-vision-instruct' },
//   { label: 'Phi-4', value: 'phi-4' },
//   { label: 'Phi-4-mini-instruct', value: 'phi-4-mini-instruct' },
//   { label: 'Phi-4-multimodal-instruct', value: 'phi-4-multimodal-instruct' },
] as const;

type Model = typeof AVAILABLE_MODELS[number]['value'];

interface Settings {
  model: Model;
  temperature: number;
}

const defaultSettings: Settings = {
  model: 'claude-3.7-sonnet',
  temperature: 0.3
};

const SettingsContext = createContext<{
  settings: Settings;
  setSettings: React.Dispatch<React.SetStateAction<Settings>>;
}>({ settings: defaultSettings, setSettings: () => {} });

export function SettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState(() => {
    const saved = localStorage.getItem('settings');
    return saved ? JSON.parse(saved) : defaultSettings;
  });

  // Save to localStorage whenever settings change
  React.useEffect(() => {
    localStorage.setItem('settings', JSON.stringify(settings));
  }, [settings]);

  return (
    <SettingsContext.Provider value={{ settings, setSettings }}>
      {children}
    </SettingsContext.Provider>
  );
}

// Hook to use settings in components
export const useSettings = () => useContext(SettingsContext);