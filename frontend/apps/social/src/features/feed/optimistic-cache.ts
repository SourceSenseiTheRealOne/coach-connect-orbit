type Cancel = () => Promise<unknown>;

export async function cancelFeedQueriesBeforeSnapshot(
  cancelLatest: Cancel,
  cancelSaved: Cancel,
): Promise<void> {
  await Promise.all([cancelLatest(), cancelSaved()]);
}
