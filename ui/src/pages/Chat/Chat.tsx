import { useState, useRef, useEffect } from 'react';
import {
    TextInput, Group, ActionIcon, Text,
    ScrollArea, Avatar, Loader, Box, Stack,
    Container,
    Flex
} from '@mantine/core';
import { IconAi, IconSend } from '@tabler/icons-react';
import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import rehypeRaw from 'rehype-raw';
import remarkGfm from 'remark-gfm';
import { ChatMessage, useChatService } from '../../services/ChatApiService';
import classes from './Chat.module.css';

interface Message {
    id: number;
    text: string;
    html?: string;
    sender: 'user' | 'ai';
    timestamp: Date;
}

export function ChatContent() {
    const [messages, setMessages] = useState<Message[]>([
        {
            id: 1,
            text: "Hello! How can I help you today?",
            sender: 'ai',
            timestamp: new Date()
        }
    ]);
    const [input, setInput] = useState('');
    const [loading, setLoading] = useState(false);
    const [animationComplete, setAnimationComplete] = useState(false);
    const [conversationHistory, setConversationHistory] = useState<ChatMessage[]>([]);

    const chatService = useChatService();

    const viewport = useRef<HTMLDivElement>(null);


    // Set animation complete after animation duration
    useEffect(() => {
        const timer = setTimeout(() => {
            setAnimationComplete(true);
        }, 2500); // 2.5s - gives enough time for the 2x 1s animations to complete

        return () => clearTimeout(timer);
    }, []);

    // Scroll to bottom whenever messages change
    useEffect(() => {
        setTimeout(() => {
            if (viewport.current) {
                viewport.current.scrollTo({ top: viewport.current.scrollHeight, behavior: 'smooth' });
            }
        }, 100);
    }, [messages]);

    const handleSend = async () => {
        if (!input.trim() || loading) return;

        const userMessage: Message = {
            id: Date.now(),
            text: input,
            sender: 'user',
            timestamp: new Date()
        };
        // Add user message to chat
        setMessages(prev => [...prev, userMessage]);

        // Store user message in conversation history
        const userChatMessage: ChatMessage = {
            role: 'user',
            content: input
        };

        const updatedHistory = [...conversationHistory, userChatMessage];
        setConversationHistory(updatedHistory);

        setInput('');
        setLoading(true);
        try {
            // Call API with the updated conversation history
            console.log('sending message:', updatedHistory);
            
            const response = await chatService.sendMessage(updatedHistory);
            // Create AI message from response
            const aiMessage: Message = {
                id: Date.now() + 1,
                text: response.content,
                sender: 'ai',
                timestamp: new Date()
            };

            // Update conversation history with AI response
            setConversationHistory(prev => [
                ...prev,
                { role: 'assistant', content: response.content }
            ]);

            setMessages(prev => [...prev, aiMessage]);
        } catch (error) {
            console.error('Failed to get AI response:', error);
            // Handle error - show error message to user
            const errorMessage: Message = {
                id: Date.now() + 1,
                text: "Sorry, I couldn't connect to the AI service. Please try again later.",
                sender: 'ai',
                timestamp: new Date()
            };
            setMessages(prev => [...prev, errorMessage]);
        } finally {
            setLoading(false);
        }
    };

    // Message component that handles markdown rendering
    const MessageContent = ({ message }: { message: Message }) => {
        if (message.sender === 'ai') {
            return (
                <div className={classes.markdownContent}>
                    <ReactMarkdown
                        remarkPlugins={[remarkGfm]}
                        rehypePlugins={[rehypeHighlight, rehypeRaw]}
                    >
                        {message.text}
                    </ReactMarkdown>
                </div>
            );
        }

        return <Text>{message.text}</Text>;
    };
    return (

        <Flex
            direction="column"
            style={{ height: '100vh' }}
            ml="80px"
        >
            <Box className={classes.chatHeader} style={{ flexShrink: 0 }}>
                <Container size="lg">
                    <div className={classes.matrixText} data-text="MATRIX">CDS/AI</div>
                </Container>
            </Box>

            <Box style={{ flex: 1, overflow: 'hidden' }} className={classes.messageArea}>
                <ScrollArea
                    viewportRef={viewport}
                    style={{ height: '100%' }}
                >
                    <Container size="lg" >
                        <Stack gap="lg" pt="lg">
                            {animationComplete && messages.map((message) => (
                                <Box
                                    key={message.id}
                                    className={message.sender === 'user' ? classes.userMessage : classes.aiMessage}
                                >
                                    <Group gap="xs" align="flex-start">
                                        {message.sender === 'ai' && (
                                            <Avatar
                                                color="blue"
                                                radius="xl"
                                                style={{
                                                    alignSelf: 'flex-start',  // Always align to the top
                                                    marginTop: '4px',         // Consistent margin from the top
                                                    flexShrink: 0             // Prevent the avatar from shrinking
                                                }}
                                            >
                                                <IconAi stroke="2" width="40" height="40" />
                                            </Avatar>
                                        )}

                                        <div className={classes.messageContent} style={{ display: 'flex', flexDirection: 'column' }}>
                                            <Text size="sm" fw={900} mb={4}>
                                                {message.sender === 'user' ? 'You' : ''}
                                            </Text>
                                            <MessageContent message={message} />
                                            <Text size="xs" c="dimmed" mt={4}>
                                                {message.timestamp.toLocaleTimeString()}
                                            </Text>
                                        </div>
                                    </Group>
                                </Box>
                            ))}

                            {loading && (
                                <Box className={classes.aiMessage}>
                                    <Group gap="xs" align="flex-start">
                                        <Avatar size="md" color="blue" radius="xl">AI</Avatar>
                                        <Loader size="sm" />
                                    </Group>
                                </Box>
                            )}
                        </Stack>
                    </Container>
                </ScrollArea>
            </Box>

            <Box className={classes.chatBox} style={{ flexShrink: 0 }}>
                <Container size="lg" >
                    <TextInput
                        size="xl"
                        radius="xl"
                        placeholder="Ask CDS/AI..."
                        className={classes.chatInput}
                        value={input}
                        onChange={(e) => setInput(e.currentTarget.value)}
                        onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && handleSend()}
                        rightSection={
                            <ActionIcon
                                color="blue"
                                onClick={handleSend}
                                disabled={!input.trim() || loading}
                                variant="filled"
                            >
                                <IconSend size="1.2rem" />
                            </ActionIcon>
                        }
                    />
                </Container>
            </Box>
        </Flex>
    );
}

export default function Chat() {
    return <ChatContent />;
}