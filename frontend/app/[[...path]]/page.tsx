import type { AsyncLocalStorage } from 'node:async_hooks';
import { notFound } from 'next/navigation';
import { Screens } from '../../screens';
import type { PageData } from '../../screens/view';

export const dynamic = 'force-dynamic';

export default async function Page() {
  const scope = Reflect.get(
    globalThis,
    Symbol.for('zentraspace.request-view'),
  ) as AsyncLocalStorage<{ view: PageData }> | undefined;
  const request = scope?.getStore();
  if (!request) notFound();
  return <Screens page={request.view} />;
}
