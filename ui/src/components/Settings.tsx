import { useDisclosure } from '@mantine/hooks';
import { Drawer, Button, Stack, Select, NumberInput } from '@mantine/core';
import { useSettings, AVAILABLE_MODELS, ModelOption } from '../context/SettingsContext';

interface SettingsDrawerProps {
    opened: boolean;
    onClose: () => void;
}

export function SettingsDrawer({ opened, onClose }: SettingsDrawerProps) {
    const { settings, setSettings } = useSettings();
    const handleModelChange = (value: string | null) => {
        if (value) {
            setSettings({
                ...settings,
                model: value as ModelOption['value']  // More explicit
            });
        }
    };
    const handleTemperatureChange = (value: string | number) => {
        if (value) {
            setSettings({
                ...settings,
                temperature: Number(value)
            });
        }
    };

    return (
        <Drawer
            opened={opened}
            onClose={onClose}
            title="Settings"
            position="left"
            overlayProps={{ backgroundOpacity: 0.5, blur: 4 }}
        >
            <Stack gap="md" p="md">
                <Select
                    label="AI MODEL"
                    placeholder="Pick value"
                    data={AVAILABLE_MODELS}
                    onChange={handleModelChange}
                    value={settings.model}
                    styles={{ label: { fontWeight: 900 } }}
                />
                <NumberInput
                    mt="md"
                    placeholder="Pick one"
                    label="AI TEMPERATURE"
                    description="Controls the randomness/creativity of the AI model's responses"
                    // classNames={classes}
                    styles={{ label: { fontWeight: 900 } }}
                    min={0}
                    max={1}
                    step={0.1}
                    onChange={handleTemperatureChange}
                    value={settings.temperature}
                />
                
                <Button mt="xl" onClick={onClose}>Save</Button>
            </Stack>
        </Drawer>
    );
}

export function modalProps() {
    const [opened, { open, close }] = useDisclosure(false);
    return { opened, open, close };
}

// Optional: Keep a standalone component for direct usage
export default function Settings() {
    const { opened, open, close } = modalProps();

    return (
        <>
            <SettingsDrawer opened={opened} onClose={close} />
            <Button variant="default" onClick={open}>
                Open Settings
            </Button>
        </>
    );
}
