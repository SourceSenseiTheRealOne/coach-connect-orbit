type Refresh = () => Promise<unknown>;

export async function refreshCommentMutationCaches(
  refreshComments: Refresh,
  refreshLatest: Refresh,
  refreshSaved: Refresh,
): Promise<void> {
  await Promise.all([refreshComments(), refreshLatest(), refreshSaved()]);
}
