import { useSettings } from '../context/SettingsContext';

interface ChatMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

interface ChatRequest {
  messages: ChatMessage[];
}

interface ChatResponse {
  content: string;
  // Add other fields if your API returns more data
}

export class ChatApiService {
  private apiUrl = 'http://127.0.0.1:3000/chat';

  async sendMessage(messages: ChatMessage[], model: string, temperature: number): Promise<ChatResponse> {
    try {
      const response = await fetch(this.apiUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          model: model,  // Use the model from settings
          temperature: temperature,
          messages: messages
        }),
      });

      if (!response.ok) {
        throw new Error(`API request failed with status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Error sending message to chat API:', error);
      throw error;
    }
  }
}

// Create a hook to use the chat service with settings
export function useChatService() {
  const { settings } = useSettings();
  const chatService = new ChatApiService();

  return {
    sendMessage: (messages: ChatMessage[]) => {
      const baseMessage: ChatMessage = {
        role: 'system',
        content: 'When providing answers, use markdown when applicable including formatting, lists, tables, codeblocks, etc.'
      };
      
      // Always include the system message at the start
      const fullMessages = [baseMessage, ...messages];
      console.log('useChatService', fullMessages, settings.model)
      return chatService.sendMessage(fullMessages, settings.model, settings.temperature);
    }
  };
}

export type { ChatMessage, ChatRequest, ChatResponse };