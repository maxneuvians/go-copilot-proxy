import { useLocation, useNavigate, Outlet } from 'react-router-dom';

import {
  IconBrandLine,
  IconHome2,
  IconInfoCircle,
  IconSettings,
} from '@tabler/icons-react';
import { Center, Stack, Tooltip, UnstyledButton, Image, ActionIcon } from '@mantine/core';

import classes from './Navbar.module.css';
import DarkModeToggle from '../../components/DarkModeToggle';
import { SettingsDrawer, modalProps } from '../../components/Settings';

interface NavbarLinkProps {
  icon: typeof IconHome2;
  label: string;
  active?: boolean;
  onClick?: () => void;
}

function NavbarLink({ icon: Icon, label, active, onClick }: NavbarLinkProps) {
  return (
    <Tooltip label={label} position="right" transitionProps={{ duration: 0 }}>
      <UnstyledButton onClick={onClick} className={classes.link} data-active={active || undefined}>
        <Icon size={20} stroke={2} />
      </UnstyledButton>
    </Tooltip>
  );
}

const mockdata = [
  { icon: IconBrandLine, label: 'New chat', url: '/chat' },
];

export default function Chat() {
  const { opened, open, close } = modalProps();

  const location = useLocation();
  const navigate = useNavigate();

  const isAboutRoute = location.pathname === '/about';
  const isChatRoute = location.pathname === '/chat';

  const links = mockdata.map((link) => (
    <NavbarLink
      {...link}
      key={link.label}
      active={isChatRoute}
      onClick={() => navigate(link.url)}
    />
  ));

  return (
    <div className={classes.appShell}>
      <SettingsDrawer opened={opened} onClose={close} />
      <nav className={classes.navbar}>
        <Center>
          <Image src="/images/cds.png" alt="CDS logo" width={50} height={50} />
        </Center>

        <div className={classes.navbarMain}>
          <Stack justify="center" gap={0}>
            {links}
            <ActionIcon
              variant="subtle"
              onClick={open}
              title="Settings"
              className={classes.link}
            >
              <IconSettings />
            </ActionIcon>
          </Stack>
        </div>

        <Stack justify="center" gap={0}>
          <NavbarLink
            icon={IconInfoCircle}
            label="About this app"
            active={isAboutRoute}
            onClick={() => navigate('/about')}
          />
          <DarkModeToggle className={classes.link} />
        </Stack>
      </nav>

      {/* This is where the child routes will be rendered */}
      <main className={classes.main}>
        <Outlet />
      </main>
    </div>
  );
}