import type { Post } from "@coach-connect/go-api-client";

export function applyOptimisticLike(post: Post, liked: boolean): Post {
  if (post.viewerHasLiked === liked) return post;
  return {
    ...post,
    viewerHasLiked: liked,
    likeCount: Math.max(0, post.likeCount + (liked ? 1 : -1)),
  };
}

export function applyOptimisticSave(post: Post, saved: boolean): Post {
  return post.viewerHasSaved === saved
    ? post
    : { ...post, viewerHasSaved: saved };
}
