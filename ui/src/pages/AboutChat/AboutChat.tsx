import { Container, Title, Paper, Flex } from '@mantine/core';
import classes from '../Chat/Chat.module.css';

export default function AboutChat() {

  return (
    <Flex
      align="start"
      justify="center"
      style={{
        minHeight: '100vh',
        paddingTop: '10vh' // This will position the content 25% down from the top
      }}
    >
      <Container size="md">
        <Paper withBorder radius="lg" shadow="lg" p="xl" pt="5rem" pb="9rem">
          <Title order={2} ta="center" mb="4rem">
            <div className={classes.matrixText} data-text="MATRIX">About AI Proxy</div>
          </Title>
        </Paper>
      </Container>
    </Flex>
  );
}