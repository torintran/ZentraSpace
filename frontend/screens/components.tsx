import { Alert, Badge, Button, Paper } from '@mantine/core';
import {
  ArrowLeft,
  ChevronDown,
  CircleAlert,
  CheckCircle2,
} from 'lucide-react';
import type { ComponentPropsWithoutRef, ReactNode } from 'react';
import { clsx } from './presentation';

export function Panel({
  as = 'section',
  ...props
}: ComponentPropsWithoutRef<'section'> & {
  as?: 'section' | 'article' | 'div';
}) {
  return <Paper component={as} shadow="xs" {...props} />;
}

export function Action({
  className,
  children,
  type = 'submit',
  ...props
}: ComponentPropsWithoutRef<'button'> & { leftSection?: ReactNode }) {
  const secondary = className?.split(' ').includes('secondary');
  return (
    <Button
      type={type}
      variant={secondary ? 'default' : 'filled'}
      className={className}
      {...props}
    >
      {children}
    </Button>
  );
}

export function ActionLink({
  className,
  children,
  href,
  ...props
}: ComponentPropsWithoutRef<'a'> & { leftSection?: ReactNode }) {
  const secondary = className?.split(' ').includes('secondary');
  return (
    <Button
      component="a"
      href={href ?? '/'}
      variant={secondary ? 'default' : 'filled'}
      className={className}
      {...props}
    >
      {children}
    </Button>
  );
}

export function BackLink({
  href,
  children,
}: {
  href: string;
  children: ReactNode;
}) {
  return (
    <a className="back" href={href}>
      <ArrowLeft aria-hidden size={18} />
      {children}
    </a>
  );
}

export function Status({
  className,
  children,
  ...props
}: ComponentPropsWithoutRef<'span'>) {
  const positive = className?.split(' ').includes('good');
  return (
    <Badge
      component="span"
      variant="light"
      size="lg"
      color={positive ? 'teal' : 'gray'}
      className={className}
      {...props}
    >
      {children}
    </Badge>
  );
}

export function Notice({
  role,
  className,
  children,
  id,
}: ComponentPropsWithoutRef<'p'>) {
  const error = role === 'alert';
  return (
    <Alert
      role={role}
      id={id}
      variant="light"
      color={error ? 'red' : 'teal'}
      icon={
        error ? (
          <CircleAlert aria-hidden size={19} />
        ) : (
          <CheckCircle2 aria-hidden size={19} />
        )
      }
      className={className}
    >
      {children}
    </Alert>
  );
}

export function FieldInput({
  className,
  placeholder,
  ...props
}: Omit<ComponentPropsWithoutRef<'input'>, 'placeholder'> & {
  placeholder?: string | number;
  'data-room-search'?: string;
}) {
  return (
    <input
      className={clsx('mantine-Input-input', className)}
      placeholder={placeholder === undefined ? undefined : String(placeholder)}
      {...props}
    />
  );
}

export function FieldArea({
  className,
  ...props
}: ComponentPropsWithoutRef<'textarea'>) {
  return (
    <textarea className={clsx('mantine-Input-input', className)} {...props} />
  );
}

export function FieldSelect(props: ComponentPropsWithoutRef<'select'>) {
  return (
    <span className="field-select">
      <select {...props} />
      <ChevronDown className="field-select-icon" size={18} aria-hidden />
    </span>
  );
}
