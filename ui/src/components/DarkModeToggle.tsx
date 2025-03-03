import { ActionIcon, useMantineColorScheme } from '@mantine/core';

import { IconMoonStars, IconSunFilled } from '@tabler/icons-react';



interface DarkModeToggleProps {
  className?: string;
}

export default function DarkModeToggle({ className }: DarkModeToggleProps) {
  const { colorScheme, setColorScheme } = useMantineColorScheme({
    keepTransitions: true,
  });
  const toggleColorScheme = () => {
    setColorScheme(colorScheme === 'dark' ? 'light' : 'dark');  
  }

  return (
    <ActionIcon
      variant="subtle"
      color={colorScheme === 'light' ? 'yellow' : 'blue'}
      onClick={toggleColorScheme}
      title="Toggle color scheme"
      className={className}
    >
      {colorScheme === 'light' ? (
        <IconMoonStars width="18" height="18" />
      ) : (
        <IconSunFilled width="18" height="18" />
      )}
    </ActionIcon>
  );
}
