"use client";

import type { Post } from "@coach-connect/go-api-client";
import { trpc } from "@coach-connect/trpc-client";
import Image from "next/image";
import { type FormEvent, useState } from "react";
import { refreshCommentMutationCaches } from "./comment-cache";
import { useFeed, type FeedView } from "./use-feed";

const control =
  "min-h-11 rounded-xl border border-slate-300 bg-white px-4 py-2 text-base outline-none focus-visible:ring-2 focus-visible:ring-lime-500 disabled:cursor-not-allowed disabled:opacity-50 dark:border-slate-700 dark:bg-slate-900";
const button = `${control} font-semibold transition hover:border-lime-500`;

function Avatar({ post }: { post: Post }) {
  const initial = post.author.displayName.slice(0, 1).toUpperCase();

  return (
    <span className="flex min-w-0 items-center gap-3">
      <span className="flex size-11 shrink-0 items-center justify-center overflow-hidden rounded-full bg-lime-200 font-bold text-slate-900">
        {post.author.avatarUrl ? (
          <Image
            alt=""
            className="size-full object-cover"
            height={44}
            src={post.author.avatarUrl}
            unoptimized
            width={44}
          />
        ) : (
          initial
        )}
      </span>
      <span className="truncate font-semibold">{post.author.displayName}</span>
    </span>
  );
}

function Comments({ post }: { post: Post }) {
  const [open, setOpen] = useState(false);
  const [body, setBody] = useState("");
  const utils = trpc.useUtils();
  const comments = trpc.social.comments.useInfiniteQuery(
    { postId: post.id, limit: 20 },
    { enabled: open, getNextPageParam: (page) => page.nextCursor },
  );
  const refresh = () => comments.refetch();
  const refreshAfterCountChange = () =>
    refreshCommentMutationCaches(
      refresh,
      () => utils.social.feed.invalidate(),
      () => utils.social.saved.invalidate(),
    );
  const create = trpc.social.createComment.useMutation({
    onSuccess: async () => {
      setBody("");
      await refreshAfterCountChange();
    },
  });
  const update = trpc.social.updateComment.useMutation({ onSuccess: refresh });
  const remove = trpc.social.deleteComment.useMutation({
    onSuccess: refreshAfterCountChange,
  });

  return (
    <div className="border-t border-slate-200 pt-3 dark:border-slate-800">
      <button
        aria-expanded={open}
        className={button}
        onClick={() => setOpen((value) => !value)}
        type="button"
      >
        {open ? "Hide comments" : `Comments (${post.commentCount})`}
      </button>
      {open ? (
        <div className="mt-4 space-y-3">
          {comments.isLoading ? <p role="status">Loading comments...</p> : null}
          {comments.isError ? (
            <button className={button} onClick={() => void refresh()}>
              Retry comments
            </button>
          ) : null}
          {comments.data?.pages
            .flatMap((page) => page.items)
            .map((comment) => (
              <article
                className="rounded-xl bg-slate-100 p-3 dark:bg-slate-800"
                key={comment.id}
              >
                <div className="flex items-center justify-between gap-3">
                  <span className="font-semibold">
                    {comment.author.displayName}
                  </span>
                  {comment.viewerOwns ? (
                    <span className="flex gap-2">
                      <button
                        className="min-h-11 px-2 underline"
                        onClick={() => {
                          const value = prompt("Edit comment", comment.body);
                          if (value?.trim()) {
                            update.mutate({
                              commentId: comment.id,
                              body: value,
                            });
                          }
                        }}
                      >
                        Edit
                      </button>
                      <button
                        className="min-h-11 px-2 underline"
                        onClick={() => {
                          if (confirm("Delete this comment?")) {
                            remove.mutate({ commentId: comment.id });
                          }
                        }}
                      >
                        Delete
                      </button>
                    </span>
                  ) : null}
                </div>
                <p className="whitespace-pre-wrap">{comment.body}</p>
              </article>
            ))}
          <form
            className="flex flex-col gap-2 sm:flex-row"
            onSubmit={(event) => {
              event.preventDefault();
              if (body.trim()) {
                create.mutate({ postId: post.id, body });
              }
            }}
          >
            <label className="sr-only" htmlFor={`comment-${post.id}`}>
              Add a comment
            </label>
            <input
              className={`${control} min-w-0 flex-1`}
              id={`comment-${post.id}`}
              maxLength={1000}
              onChange={(event) => setBody(event.target.value)}
              placeholder="Add a football comment"
              value={body}
            />
            <button
              className={button}
              disabled={!body.trim() || create.isPending}
            >
              Comment
            </button>
          </form>
          {comments.hasNextPage ? (
            <button
              className={button}
              disabled={comments.isFetchingNextPage}
              onClick={() => void comments.fetchNextPage()}
            >
              Load more comments
            </button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function Card({
  post,
  feed,
}: {
  post: Post;
  feed: ReturnType<typeof useFeed>;
}) {
  return (
    <article className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900 sm:p-6">
      <header className="flex items-start justify-between gap-4">
        <button
          aria-label={`Profile for ${post.author.displayName} (coming later)`}
          className="min-h-11 min-w-0 text-left"
          type="button"
        >
          <Avatar post={post} />
        </button>
        {post.viewerOwns ? (
          <span className="flex gap-1">
            <button
              className="min-h-11 px-3 underline"
              onClick={() => {
                const body = prompt("Edit post", post.body);
                if (body?.trim()) {
                  feed.update.mutate({ postId: post.id, body });
                }
              }}
            >
              Edit
            </button>
            <button
              className="min-h-11 px-3 underline"
              onClick={() => {
                if (confirm("Delete this post?")) {
                  feed.remove.mutate({ postId: post.id });
                }
              }}
            >
              Delete
            </button>
          </span>
        ) : null}
      </header>
      <p className="my-4 max-w-prose whitespace-pre-wrap text-base leading-7">
        {post.body}
      </p>
      {post.media ? (
        <Image
          alt={post.media.altText}
          className="mb-4 aspect-video w-full rounded-xl object-cover"
          height={post.media.height}
          src={post.media.url}
          unoptimized
          width={post.media.width}
        />
      ) : null}
      <div className="mb-4 flex flex-wrap gap-2">
        <button
          aria-pressed={post.viewerHasLiked}
          className={button}
          disabled={feed.like.isPending}
          onClick={() =>
            feed.like.mutate({
              postId: post.id,
              liked: !post.viewerHasLiked,
            })
          }
        >
          {post.viewerHasLiked ? "Unlike" : "Like"} - {post.likeCount}
        </button>
        <button
          aria-pressed={post.viewerHasSaved}
          className={button}
          disabled={feed.save.isPending}
          onClick={() =>
            feed.save.mutate({
              postId: post.id,
              saved: !post.viewerHasSaved,
            })
          }
        >
          {post.viewerHasSaved ? "Unsave" : "Save"}
        </button>
      </div>
      <Comments post={post} />
    </article>
  );
}

export function Feed() {
  const [view, setView] = useState<FeedView>("latest");
  const [body, setBody] = useState("");
  const feed = useFeed(view);

  function submit(event: FormEvent) {
    event.preventDefault();
    if (!body.trim()) {
      return;
    }
    feed.create.mutate({ body }, { onSuccess: () => setBody("") });
  }

  return (
    <main className="min-h-screen bg-slate-50 px-4 py-6 dark:bg-slate-950 sm:px-6 lg:py-10">
      <div className="mx-auto w-full max-w-2xl space-y-5">
        <header>
          <p className="text-sm font-bold uppercase tracking-widest text-lime-700 dark:text-lime-400">
            Matchday network
          </p>
          <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
            Football feed
          </h1>
        </header>
        <nav aria-label="Feed views" className="flex gap-2">
          <button
            aria-current={view === "latest" ? "page" : undefined}
            className={button}
            onClick={() => setView("latest")}
          >
            Latest
          </button>
          <button
            aria-current={view === "saved" ? "page" : undefined}
            className={button}
            onClick={() => setView("saved")}
          >
            Saved
          </button>
        </nav>
        {view === "latest" ? (
          <form
            className="rounded-2xl border border-slate-200 bg-white p-4 dark:border-slate-800 dark:bg-slate-900"
            onSubmit={submit}
          >
            <label className="mb-2 block font-semibold" htmlFor="post-body">
              Share a football insight
            </label>
            <textarea
              className={`${control} min-h-28 w-full resize-y`}
              id="post-body"
              maxLength={2000}
              onChange={(event) => setBody(event.target.value)}
              placeholder="What did you notice in training or on matchday?"
              value={body}
            />
            <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
              <button
                className={button}
                disabled
                title="Secure image upload is not available yet"
                type="button"
              >
                Add image (coming later)
              </button>
              <button
                className={button}
                disabled={!body.trim() || feed.create.isPending}
              >
                {feed.create.isPending ? "Posting..." : "Post insight"}
              </button>
            </div>
            {feed.create.error ? (
              <p className="mt-2 text-red-700 dark:text-red-300" role="alert">
                Your post was not published. Try again.
              </p>
            ) : null}
          </form>
        ) : null}
        {feed.query.isLoading ? (
          <div className="space-y-3" role="status">
            <div className="h-48 animate-pulse rounded-2xl bg-slate-200 dark:bg-slate-800" />
            <span className="sr-only">Loading feed</span>
          </div>
        ) : null}
        {feed.query.isError ? (
          <section
            className="rounded-2xl border border-red-300 p-5"
            role="alert"
          >
            <p>We could not load the feed.</p>
            <button
              className={`${button} mt-3`}
              onClick={() => void feed.query.refetch()}
            >
              Retry
            </button>
          </section>
        ) : null}
        {!feed.query.isLoading &&
        !feed.query.isError &&
        feed.posts.length === 0 ? (
          <section className="rounded-2xl border border-dashed border-slate-300 p-8 text-center dark:border-slate-700">
            <h2 className="text-xl font-semibold">
              {view === "saved"
                ? "No saved posts yet"
                : "The touchline is quiet"}
            </h2>
            <p className="mt-2 text-slate-600 dark:text-slate-300">
              {view === "saved"
                ? "Save useful football insights to find them here privately."
                : "Be the first to share a football insight."}
            </p>
          </section>
        ) : null}
        <section
          aria-label={view === "saved" ? "Saved posts" : "Latest posts"}
          className="space-y-4"
        >
          {feed.posts.map((post) => (
            <Card feed={feed} key={post.id} post={post} />
          ))}
        </section>
        {feed.query.hasNextPage ? (
          <button
            className={`${button} w-full`}
            disabled={feed.query.isFetchingNextPage}
            onClick={() => void feed.query.fetchNextPage()}
          >
            {feed.query.isFetchingNextPage ? "Loading..." : "Load more posts"}
          </button>
        ) : null}
      </div>
    </main>
  );
}
