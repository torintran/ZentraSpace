export function clsx(...classes: Array<string | undefined>) {
  return classes.filter(Boolean).join(' ');
}
