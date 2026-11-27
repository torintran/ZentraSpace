'use client';

import { createTheme, MantineProvider } from '@mantine/core';
import type { ReactNode } from 'react';

const theme = createTheme({
  primaryColor: 'teal',
  primaryShade: 8,
  defaultRadius: 'md',
  fontFamily: 'system-ui, "Segoe UI", sans-serif',
  headings: { fontFamily: 'system-ui, "Segoe UI", sans-serif' },
  components: {
    Button: { defaultProps: { size: 'md' } },
    Paper: { defaultProps: { radius: 'lg', withBorder: true } },
  },
});

export function Providers({ children }: { children: ReactNode }) {
  return (
    <MantineProvider theme={theme} forceColorScheme="light">
      {children}
    </MantineProvider>
  );
}
