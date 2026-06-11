# Coub.com API v2 — Comprehensive Reference

**Base URL:** `https://coub.com/api/v2`

All endpoints are prefixed with `/api/v2` unless stated otherwise.  
Sources: reverse-engineered from [Derfirm/coub_api](https://github.com/Derfirm/coub_api),
[Bukk94/CoubDownloader](https://github.com/Bukk94/CoubDownloader), and
[HelpSeeker/CoubDownloader](https://github.com/HelpSeeker/CoubDownloader).

---

## Table of Contents

1. [Authentication](#1-authentication)
2. [Coubs](#2-coubs)
3. [Timelines](#3-timelines)
4. [Search](#4-search)
5. [Channels](#5-channels)
6. [Users](#6-users)
7. [Likes](#7-likes)
8. [Recoubs](#8-recoubs)
9. [Follows](#9-follows)
10. [Friends](#10-friends)
11. [Notifications](#11-notifications)
12. [Action Subject Metadata](#12-action-subject-metadata)
13. [Data Structures](#13-data-structures)
14. [Enumerated Values](#14-enumerated-values)
15. [Notes and Limitations](#15-notes-and-limitations)

---

## 1. Authentication

Coub uses token-based authentication. The token is passed as a query parameter or
via a cookie header, depending on the client implementation.

### Query parameter (OAuth-style access token)

```
?access_token=<token>
```

Append to any authenticated endpoint.

### Cookie header (session token)

```
Cookie: remember_token=<token>
```

Used by browser-based clients and some downloader tools. The `remember_token` value
can be extracted from the browser's cookie storage after logging in at coub.com.

### Which endpoints require authentication

Endpoints marked **Auth required** below will return `403 Forbidden` without a
valid token. All other endpoints are publicly accessible.

---

## 2. Coubs

### 2.1 Get a single coub

```
GET /api/v2/coubs/{permalink}
```

Retrieves the full metadata object for one coub.

**Path parameters**

| Parameter   | Type   | Description                                        |
|-------------|--------|----------------------------------------------------|
| `permalink` | string | The coub identifier — the short string in the URL: `https://coub.com/view/{permalink}` |

**Authentication:** Not required (public coubs). Some fields are only present for
authenticated requests (see [BigCoub](#bigcoub)).

**Response:** [`BigCoub`](#bigcoub) object.

**Example request**

```
GET https://coub.com/api/v2/coubs/2iq3e4
```

---

### 2.2 Get coub segments

```
GET /api/v2/coubs/{permalink}/segments
```

Returns the loop/segment timing data for a coub. Not all coubs have segments;
the endpoint returns `404` for coubs without segment data (common for recoubs).

**Path parameters**

| Parameter   | Type   | Description      |
|-------------|--------|------------------|
| `permalink` | string | Coub identifier  |

**Authentication:** Not required.

---

### 2.3 Initialise a new upload

```
POST /api/v2/coubs/init_upload
```

Creates a new empty coub object server-side and returns its numeric ID. This ID
is required for the subsequent upload steps.

**Authentication:** Required.

**Response fields**

| Field | Type | Description                         |
|-------|------|-------------------------------------|
| `id`  | int  | Numeric ID of the new coub draft    |

---

### 2.4 Upload video stream

```
POST /api/v2/coubs/{coub_id}/upload_video
```

Uploads raw video data for a coub draft. The request body is the raw binary of
the video file.

**Path parameters**

| Parameter | Type | Description                          |
|-----------|------|--------------------------------------|
| `coub_id` | int  | Numeric coub ID from `init_upload`   |

**Headers**

| Header         | Value                                        |
|----------------|----------------------------------------------|
| `Content-Type` | MIME type of the video (e.g. `video/mp4`)    |

**Authentication:** Required.

**Request body:** Raw binary video file.

---

### 2.5 Upload audio stream

```
POST /api/v2/coubs/{coub_id}/upload_audio
```

Uploads raw audio data for a coub draft.

**Path parameters**

| Parameter | Type | Description                        |
|-----------|------|------------------------------------|
| `coub_id` | int  | Numeric coub ID from `init_upload` |

**Headers**

| Header         | Value                                       |
|----------------|---------------------------------------------|
| `Content-Type` | MIME type of the audio (e.g. `audio/mpeg`) |

**Authentication:** Required.

**Request body:** Raw binary audio file.

---

### 2.6 Finalise upload (publish)

```
POST /api/v2/coubs/{coub_id}/finalize_upload
```

Triggers server-side processing and optionally publishes the coub.

**Path parameters**

| Parameter | Type | Description                        |
|-----------|------|------------------------------------|
| `coub_id` | int  | Numeric coub ID from `init_upload` |

**Query / body parameters**

| Parameter                   | Type    | Required | Description                                          |
|-----------------------------|---------|----------|------------------------------------------------------|
| `title`                     | string  | Yes      | Display title for the coub                          |
| `tags`                      | string  | No       | Comma-separated list of tag strings                 |
| `original_visibility_type`  | string  | No       | One of `public`, `friends`, `unlisted`, `private` (default: `private`) |
| `sound_enabled`             | boolean | No       | Whether audio is enabled (default: `true`)          |

**Authentication:** Required.

---

### 2.7 Get upload/processing status

```
GET /api/v2/coubs/{coub_id}/finalize_status
```

Polls the processing state of a coub that was just finalised.

**Path parameters**

| Parameter | Type | Description       |
|-----------|------|-------------------|
| `coub_id` | int  | Numeric coub ID   |

**Authentication:** Required.

---

### 2.8 Edit coub metadata

```
POST /api/v2/coubs/{permalink}/update_info
```

Updates the title, tags, and visibility of an existing coub.

**Path parameters**

| Parameter   | Type   | Description      |
|-------------|--------|------------------|
| `permalink` | string | Coub identifier  |

**Query / body parameters**

| Parameter                          | Type   | Required | Description                              |
|------------------------------------|--------|----------|------------------------------------------|
| `coub[channel_id]`                 | int    | Yes      | ID of the owning channel                |
| `coub[title]`                      | string | Yes      | New title                                |
| `coub[tags]`                       | string | No       | Comma-separated tag list                 |
| `coub[original_visibility_type]`   | string | No       | Visibility: `public`, `friends`, `unlisted`, `private` |

**Authentication:** Required (must be the coub owner).

**Response:** [`BigCoub`](#bigcoub) object.

---

## 3. Timelines

Timelines return paginated lists of coubs from a variety of sources.

### Common pagination parameters (all timeline endpoints)

| Parameter  | Type | Default | Description                             |
|------------|------|---------|-----------------------------------------|
| `page`     | int  | 1       | Page number (1-indexed)                 |
| `per_page` | int  | 25      | Coubs per page (max 25)                 |

### Common response envelope (all timeline endpoints)

```json
{
  "page": 1,
  "per_page": 25,
  "total_pages": 42,
  "coubs": [ /* array of BigCoub objects */ ]
}
```

---

### 3.1 Hot section / subscriptions timeline

```
GET /api/v2/timeline/subscriptions/{period_or_section}
```

Returns coubs from Coub's "hot" feed. The `{period_or_section}` path segment
determines the sort window.

**Path segment values**

| Value       | Description                         |
|-------------|-------------------------------------|
| `daily`     | Hot — last 24 hours                 |
| `weekly`    | Hot — last 7 days                   |
| `monthly`   | Hot — last 30 days (default on site)|
| `quarter`   | Hot — last 3 months                 |
| `half`      | Hot — last 6 months                 |
| `rising`    | Rising coubs                        |
| `fresh`     | Newest coubs                        |

**Query parameters**

| Parameter  | Type   | Default           | Description                                                        |
|------------|--------|-------------------|--------------------------------------------------------------------|
| `order_by` | string | `newest_popular`  | Sort order within the period. Values: `likes_count`, `views_count`, `newest_popular`, `oldest` |
| `page`     | int    | 1                 | Page number                                                        |
| `per_page` | int    | 25                | Items per page (max 25)                                            |

**Authentication:** Not required.

**Notes:** The API caps this timeline at 99 pages.

**Example**

```
GET https://coub.com/api/v2/timeline/subscriptions/monthly?per_page=25&page=1
```

---

### 3.2 Community timeline

```
GET /api/v2/timeline/community/{category}/{period_or_section}
```

Returns coubs from a specific community/category.

**Path parameters**

| Parameter            | Type   | Description                                                         |
|----------------------|--------|---------------------------------------------------------------------|
| `category`           | string | Community permalink (see [Category values](#category))              |
| `period_or_section`  | string | Time window or section (same values as §3.1 path segment)           |

For `top` and `views_count` sort orders the path becomes:

```
GET /api/v2/timeline/community/{category}/fresh?order_by={likes_count|views_count}
```

For the `random` sort order:

```
GET /api/v2/timeline/random/{category}
```

**Query parameters**

| Parameter  | Type | Default | Description      |
|------------|------|---------|------------------|
| `page`     | int  | 1       | Page number      |
| `per_page` | int  | 25      | Items per page   |

**Authentication:** Not required.

**Notes:** Capped at 99 pages.

**Example**

```
GET https://coub.com/api/v2/timeline/community/animals-pets/monthly?per_page=25&page=1
```

---

### 3.3 Channel timeline

```
GET /api/v2/timeline/channel/{channel_permalink}
```

Returns coubs uploaded by (or recoubbed to) a specific channel.

**Path parameters**

| Parameter           | Type   | Description                                 |
|---------------------|--------|---------------------------------------------|
| `channel_permalink` | string | The channel's URL slug (e.g. `some-channel`)|

**Query parameters**

| Parameter  | Type   | Default       | Description                                                                            |
|------------|--------|---------------|----------------------------------------------------------------------------------------|
| `order_by` | string | `newest`      | Sort order. Values: `newest` (alias `date`), `likes_count`, `views_count`, `oldest`, `random` |
| `type`     | string | *(both)*      | Filter by coub type. Values: `simples` (originals only), `recoubs` (recoubs only). Omit for both. |
| `page`     | int    | 1             | Page number                                                                            |
| `per_page` | int    | 25            | Items per page (max 25)                                                                |

**Authentication:** Not required.

**Example**

```
GET https://coub.com/api/v2/timeline/channel/some-channel?per_page=25&order_by=newest&page=1
```

---

### 3.4 Tag timeline

```
GET /api/v2/timeline/tag/{tag_name}
```

Returns coubs associated with a tag.

**Path parameters**

| Parameter  | Type   | Description                           |
|------------|--------|---------------------------------------|
| `tag_name` | string | URL-encoded tag string                |

**Query parameters**

| Parameter  | Type   | Default            | Description                                                            |
|------------|--------|--------------------|------------------------------------------------------------------------|
| `order_by` | string | `newest_popular`   | Sort order. Values: `newest_popular`, `likes_count`, `views_count`, `newest` |
| `page`     | int    | 1                  | Page number                                                            |
| `per_page` | int    | 25                 | Items per page (max 25)                                                |

**Authentication:** Not required.

**Notes:** Capped at 99 pages by the API (requests for page > 99 redirect back to page 1).

**Example**

```
GET https://coub.com/api/v2/timeline/tag/cats?per_page=25&order_by=newest_popular&page=1
```

---

### 3.5 Featured / Explore timeline

```
GET /api/v2/timeline/explore
```

Returns editorially curated (featured) coubs.

**Query parameters**

| Parameter  | Type   | Default  | Description                                                                  |
|------------|--------|----------|------------------------------------------------------------------------------|
| `order_by` | string | *(none)* | Section. Values: *(empty — recent/default)*, `top_of_the_month`, `undervalued` |
| `page`     | int    | 1        | Page number                                                                  |
| `per_page` | int    | 25       | Items per page                                                               |

**Authentication:** Not required.

---

### 3.6 Coub of the Day timeline

```
GET /api/v2/timeline/explore/coub_of_the_day
```

Returns the "Coub of the Day" selection.

**Query parameters**

| Parameter  | Type   | Default  | Description                                              |
|------------|--------|----------|----------------------------------------------------------|
| `order_by` | string | *(none)* | Values: *(empty — recent/default)*, `top`, `views_count` |
| `page`     | int    | 1        | Page number                                              |
| `per_page` | int    | 25       | Items per page                                           |

**Authentication:** Not required.

---

### 3.7 Random timeline

```
GET /api/v2/timeline/explore/random
```

Returns a random selection of coubs.

**Query parameters**

| Parameter  | Type   | Default  | Description                          |
|------------|--------|----------|--------------------------------------|
| `order_by` | string | *(none)* | Values: *(empty — popular)*, `top`   |
| `page`     | int    | 1        | Page number                          |
| `per_page` | int    | 25       | Items per page                       |

**Authentication:** Not required.

---

### 3.8 Random community timeline

```
GET /api/v2/timeline/random/{category}
```

Returns random coubs from a specific community.

**Path parameters**

| Parameter  | Type   | Description                                             |
|------------|--------|---------------------------------------------------------|
| `category` | string | Community permalink (see [Category values](#category)) |

**Query parameters**

| Parameter  | Type | Default | Description    |
|------------|------|---------|----------------|
| `page`     | int  | 1       | Page number    |
| `per_page` | int  | 25      | Items per page |

**Authentication:** Not required.

---

### 3.9 Authenticated user feed (own timeline)

```
GET /api/v2/timeline
```

Returns the personalised home feed for the authenticated user.

**Query parameters**

| Parameter  | Type | Default | Description    |
|------------|------|---------|----------------|
| `page`     | int  | 1       | Page number    |
| `per_page` | int  | 25      | Items per page |

**Authentication:** Required.

**Response:** `MyTimeLineResponse` envelope (same structure as standard timeline).

---

### 3.10 Authenticated user liked coubs

```
GET /api/v2/timeline/likes
```

Returns coubs liked by the authenticated user.

**Query parameters**

| Parameter  | Type   | Default | Description                                                       |
|------------|--------|---------|-------------------------------------------------------------------|
| `order_by` | string | *(none)*| Sort order. Values: `oldest`, `likes_count`, `views_count`, `date` |
| `all`      | bool   | false   | When `true`, returns aggregate count rather than full listing (used internally to get page counts) |
| `page`     | int    | 1       | Page number                                                       |
| `per_page` | int    | 25      | Items per page (max 25)                                           |

**Authentication:** Required (cookie `remember_token` or `access_token` query param).

**Notes:** The API caps this at 999 pages.

---

### 3.11 Authenticated user bookmarks (favourites)

```
GET /api/v2/timeline/favourites
```

Returns coubs bookmarked (saved/favourited) by the authenticated user.

**Query parameters**

| Parameter  | Type   | Default | Description                                               |
|------------|--------|---------|-----------------------------------------------------------|
| `order_by` | string | *(none)*| Sort order. Values: `oldest`, `likes_count`, `views_count`, `date` |
| `page`     | int    | 1       | Page number                                               |
| `per_page` | int    | 25      | Items per page (max 25)                                   |

**Authentication:** Required.

---

## 4. Search

### 4.1 General search (coubs + channels)

```
GET /api/v2/search
```

Returns coubs and channels matching the search query.

**Query parameters**

| Parameter  | Type   | Required | Description                                                                        |
|------------|--------|----------|------------------------------------------------------------------------------------|
| `q`        | string | Yes      | Search term                                                                        |
| `order_by` | string | No       | Sort order. Values: `newest_popular` (default), `likes_count`, `views_count`, `newest`, `oldest` |
| `page`     | int    | No       | Page number (default: 1)                                                           |
| `per_page` | int    | No       | Items per page (default: 25, max 25)                                               |

**Authentication:** Not required.

**Response**

```json
{
  "page": 1,
  "per_page": 25,
  "total_pages": 10,
  "coubs": [ /* BigCoub array */ ],
  "channels": [ /* ChannelBig array */ ]
}
```

---

### 4.2 Coub-only search

```
GET /api/v2/search/coubs
```

Returns only coubs matching the search query.

**Query parameters**

| Parameter  | Type   | Required | Description                                                                           |
|------------|--------|----------|---------------------------------------------------------------------------------------|
| `q`        | string | Yes      | Search term                                                                           |
| `order_by` | string | No       | Sort order. Values: `newest` (default), `likes_count`, `views_count`, `most_recent`  |
| `page`     | int    | No       | Page number (default: 1)                                                              |
| `per_page` | int    | No       | Items per page (default: 25, max 25)                                                  |

**Authentication:** Not required.

**Response**

```json
{
  "page": 1,
  "per_page": 25,
  "total_pages": 10,
  "coubs": [ /* BigCoub array */ ]
}
```

---

### 4.3 Channel-only search

```
GET /api/v2/search/channels
```

Returns only channels matching the search query.

**Query parameters**

| Parameter  | Type   | Required | Description                                                     |
|------------|--------|----------|-----------------------------------------------------------------|
| `q`        | string | Yes      | Search term                                                     |
| `order_by` | string | No       | Sort order. Values: `newest` (default), `followers_count`       |
| `page`     | int    | No       | Page number (default: 1)                                        |
| `per_page` | int    | No       | Items per page (default: 25, max 25)                            |

**Authentication:** Not required.

**Response**

```json
{
  "page": 1,
  "per_page": 25,
  "total_pages": 5,
  "channels": [ /* ChannelBig array */ ]
}
```

---

## 5. Channels

### 5.1 Get channel by ID

```
GET /api/v2/channels/{channel_id}
```

Returns the full profile of a channel.

**Path parameters**

| Parameter    | Type           | Description                          |
|--------------|----------------|--------------------------------------|
| `channel_id` | int or string  | Numeric channel ID or permalink slug |

**Authentication:** Not required (some fields only present when authenticated).

**Response:** [`ChannelBig`](#channelbig) object.

---

### 5.2 Create channel

```
POST /api/v2/channels
```

Creates a new channel for the authenticated user.

**Query / body parameters**

| Parameter              | Type   | Required | Description                              |
|------------------------|--------|----------|------------------------------------------|
| `channels[title]`      | string | Yes      | Display name of the channel              |
| `channels[permalink]`  | string | Yes      | URL slug (min 8 characters)              |
| `channels[category]`   | string | Yes      | Category permalink (see [Category](#category)) |

**Authentication:** Required.

**Response:** `ChannelResponse` — contains `redirect_url` (string) and `channel` ([`ChannelBig`](#channelbig)).

---

### 5.3 Delete channel

```
DELETE /api/v2/channels/{channel_id}
```

Deletes a channel owned by the authenticated user.

**Path parameters**

| Parameter    | Type | Description       |
|--------------|------|-------------------|
| `channel_id` | int  | Numeric channel ID|

**Authentication:** Required.

---

### 5.4 Update channel info

```
PUT /api/v2/channels/update_info
```

Updates metadata for the authenticated user's current channel. Exact parameter
names follow the same `channels[field]` convention as channel creation.

**Authentication:** Required.

> **Note:** This endpoint's full parameter set is undocumented in the reference
> implementations; the implementation was not completed in the source.

---

### 5.5 Upload channel avatar

```
POST /api/v2/channels/upload_avatar
```

Sets a new avatar image for a channel.

**Query parameters**

| Parameter      | Type | Required | Description       |
|----------------|------|----------|-------------------|
| `channels[id]` | int  | Yes      | Target channel ID |

**Form data**

| Field              | Type | Description            |
|--------------------|------|------------------------|
| `channels[avatar]` | file | Image file to upload   |

**Authentication:** Required.

---

### 5.6 Delete channel avatar

```
DELETE /api/v2/channels/delete_avatar
```

Removes the avatar from a channel.

**Query parameters**

| Parameter      | Type | Required | Description       |
|----------------|------|----------|-------------------|
| `channels[id]` | int  | Yes      | Target channel ID |

**Authentication:** Required.

---

### 5.7 Add channel background

```
POST /api/v2/channels/{channel_id}/backgrounds
```

Sets a background image (from a coub or an uploaded image) for a channel.

**Path parameters**

| Parameter    | Type | Description       |
|--------------|------|-------------------|
| `channel_id` | int  | Target channel ID |

**Query / body parameters** *(one of the two must be provided)*

| Parameter           | Type   | Description                              |
|---------------------|--------|------------------------------------------|
| `background[coub]`  | string | Permalink of a coub to use as background |
| *(image file)*      | file   | Uploaded image file                      |

**Authentication:** Required.

---

### 5.8 Change channel background position

```
POST /api/v2/channels/{channel_id}/backgrounds
```

Adjusts the vertical position offset of the channel background.

**Path parameters**

| Parameter    | Type | Description       |
|--------------|------|-------------------|
| `channel_id` | int  | Target channel ID |

**Query / body parameters**

| Parameter  | Type  | Description                            |
|------------|-------|----------------------------------------|
| `offset_y` | float | Vertical offset value for the image    |

**Authentication:** Required.

---

### 5.9 Delete channel background

```
DELETE /api/v2/channels/{channel_id}/backgrounds
```

Removes the background from a channel.

**Path parameters**

| Parameter    | Type | Description       |
|--------------|------|-------------------|
| `channel_id` | int  | Target channel ID |

**Authentication:** Required.

---

## 6. Users

### 6.1 Get current user profile

```
GET /api/v2/users/me
```

Returns the full profile of the authenticated user, including all their channels.

**Authentication:** Required.

**Response fields**

| Field                          | Type            | Description                                     |
|--------------------------------|-----------------|-------------------------------------------------|
| `id`                           | int             | User ID                                         |
| `permalink`                    | string          | URL slug                                        |
| `name`                         | string          | Display name                                    |
| `sex`                          | string          | `male`, `female`, or `unspecified`              |
| `city`                         | string or null  | User's city                                     |
| `current_channel`              | ChannelSmall    | Currently active channel                        |
| `created_at`                   | datetime        | Account creation timestamp (ISO 8601)           |
| `updated_at`                   | datetime        | Last update timestamp (ISO 8601)                |
| `api_token`                    | string          | The user's API token (returned on /users/me)    |
| `has_linked_vine_accounts`     | bool            | Legacy Vine link status                         |
| `likes_count`                  | int             | Total liked coubs                               |
| `favourites_count`             | int             | Total bookmarked coubs                          |
| `channels`                     | ChannelBig[]    | All channels owned by this user                 |

---

### 6.2 Switch active channel

```
PUT /api/v2/users/change_channel
```

Changes the active/current channel for the authenticated user.

**Query / body parameters**

| Parameter    | Type | Required | Description                |
|--------------|------|----------|----------------------------|
| `channel_id` | int  | Yes      | ID of the channel to switch to |

**Authentication:** Required.

---

## 7. Likes

### 7.1 Like a coub

```
POST /api/v2/likes
```

Adds a like to a coub from a specified channel.

**Query / body parameters**

| Parameter    | Type | Required | Description                                |
|--------------|------|----------|--------------------------------------------|
| `id`         | int  | Yes      | Numeric ID of the coub to like             |
| `channel_id` | int  | Yes      | Numeric ID of the channel performing the like |

**Authentication:** Required.

---

### 7.2 Unlike a coub

```
DELETE /api/v2/likes
```

Removes a like from a coub.

**Query / body parameters**

| Parameter    | Type | Required | Description                                   |
|--------------|------|----------|-----------------------------------------------|
| `id`         | int  | Yes      | Numeric ID of the coub                        |
| `channel_id` | int  | Yes      | Numeric ID of the channel removing the like   |

**Authentication:** Required.

---

## 8. Recoubs

### 8.1 Create a recoub

```
POST /api/v2/recoubs
```

Recoubs (reposts) an existing coub to a specified channel.

**Query / body parameters**

| Parameter       | Type | Required | Description                                    |
|-----------------|------|----------|------------------------------------------------|
| `recoub_to_id`  | int  | Yes      | Numeric ID of the coub being recoubbed         |
| `channel_id`    | int  | Yes      | Numeric ID of the channel performing the recoub|

**Authentication:** Required.

**Response:** [`BigCoub`](#bigcoub) object of the new recoub.

---

### 8.2 Delete a recoub

```
DELETE /api/v2/recoubs
```

Removes a recoub.

**Query / body parameters**

| Parameter    | Type | Required | Description                                  |
|--------------|------|----------|----------------------------------------------|
| `id`         | int  | Yes      | Numeric ID of the recoub to delete           |
| `channel_id` | int  | Yes      | Numeric ID of the channel that owns the recoub |

**Authentication:** Required.

---

## 9. Follows

### 9.1 Follow a channel

```
POST /api/v2/follows
```

Follows another channel.

**Query / body parameters**

| Parameter    | Type | Required | Description                                |
|--------------|------|----------|--------------------------------------------|
| `id`         | int  | Yes      | Numeric ID of the channel to follow        |
| `channel_id` | int  | Yes      | Numeric ID of the authenticated user's channel |

**Authentication:** Required.

---

### 9.2 Unfollow a channel

```
DELETE /api/v2/follows
```

Unfollows a channel.

**Query / body parameters**

| Parameter    | Type | Required | Description                                   |
|--------------|------|----------|-----------------------------------------------|
| `id`         | int  | Yes      | Numeric ID of the channel to unfollow         |
| `channel_id` | int  | Yes      | Numeric ID of the authenticated user's channel|

**Authentication:** Required.

---

## 10. Friends

### 10.1 List friends

```
GET /api/v2/friends
```

Lists channels that the authenticated user is connected to as friends via a
social provider.

**Query parameters**

| Parameter  | Type   | Default | Description                                                           |
|------------|--------|---------|-----------------------------------------------------------------------|
| `provider` | string | *(all)* | Filter by social provider. Values: `facebook`, `twitter`, `vkontakte`, `google` |
| `page`     | int    | 1       | Page number                                                           |
| `per_page` | int    | 10      | Items per page                                                        |

**Authentication:** Required.

**Response**

```json
{
  "page": 1,
  "total_pages": 3,
  "total_friends": 60,
  "per_page": 10,
  "friends": [ /* ChannelSmall array */ ]
}
```

---

### 10.2 Get recommended friends to follow

```
GET /api/v2/friends/recommended
```

Returns channel recommendations based on a search term.

**Query parameters**

| Parameter  | Type   | Required | Description                   |
|------------|--------|----------|-------------------------------|
| `q`        | string | Yes      | Search term                   |
| `page`     | int    | No       | Page number (default: 1)      |
| `per_page` | int    | No       | Items per page (default: 10)  |

**Authentication:** Required.

**Response**

```json
{
  "channels": [ /* ChannelSmall array */ ]
}
```

---

### 10.3 Get friends to follow (initial suggestion list)

```
GET /api/v2/friends/friends_to_follow
```

Returns a list of suggested channels to follow.

**Query parameters**

| Parameter | Type   | Default         | Description                                   |
|-----------|--------|-----------------|-----------------------------------------------|
| `type`    | string | `initial.json`  | Request type. Values: `initial.json`, `next`  |
| `count`   | int    | 20              | Number of suggestions to return               |

**Authentication:** Required.

**Response**

```json
{
  "friends": [ /* ChannelSmall array */ ]
}
```

---

## 11. Notifications

### 11.1 List notifications

```
GET /api/v2/notifications
```

Returns the authenticated user's notification list.

**Query parameters**

| Parameter  | Type | Default | Description    |
|------------|------|---------|----------------|
| `page`     | int  | 1       | Page number    |
| `per_page` | int  | 10      | Items per page |

**Authentication:** Required.

---

### 11.2 Mark notifications as viewed

```
GET /api/v2/channels/notifications_viewed
```

Marks all notifications as read. Requires an `access_token` header or param.

**Authentication:** Required.

> **Note:** This endpoint was not fully implemented in the reference libraries.

---

## 12. Action Subject Metadata

These endpoints return lists of channels that performed a specific action on a
coub or channel object (e.g. who liked a coub, who follows a channel).

### 12.1 Coub likes list

```
GET /api/v2/action_subjects_data/coub_likes_list
```

Returns channels that liked a specific coub.

**Query parameters**

| Parameter | Type | Required | Description          |
|-----------|------|----------|----------------------|
| `id`      | int  | Yes      | Numeric coub ID      |
| `page`    | int  | No       | Page number (def: 1) |

**Authentication:** Required.

**Response:** `MetaResponse` — `{ page, total_pages, channels: ChannelSmall[] }`

---

### 12.2 Coub recoubs list

```
GET /api/v2/action_subjects_data/recoubs_list
```

Returns channels that recoubbed a specific coub.

**Query parameters**

| Parameter | Type  | Required | Description                                |
|-----------|-------|----------|--------------------------------------------|
| `id`      | int   | Yes      | Numeric coub ID                            |
| `page`    | int   | No       | Page number (default: 1)                   |
| `ids[]`   | int   | No       | Optional: filter to a single channel ID (max 1) |

**Authentication:** Required.

**Response:** `MetaResponse` — `{ page, total_pages, channels: ChannelSmall[] }`

---

### 12.3 Channel followers list

```
GET /api/v2/action_subjects_data/followers_list
```

Returns channels that follow a specific channel.

**Query parameters**

| Parameter | Type | Required | Description                                   |
|-----------|------|----------|-----------------------------------------------|
| `id`      | int  | Yes      | Numeric channel ID                            |
| `page`    | int  | No       | Page number (default: 1)                      |
| `ids[]`   | int  | No       | Optional: filter to a single channel ID (max 1) |

**Authentication:** Required.

**Response:** `MetaResponse` — `{ page, total_pages, channels: ChannelSmall[] }`

---

### 12.4 Channel followings list

```
GET /api/v2/action_subjects_data/followings_list
```

Returns channels that a specific channel follows.

**Query parameters**

| Parameter | Type | Required | Description        |
|-----------|------|----------|--------------------|
| `id`      | int  | Yes      | Numeric channel ID |
| `page`    | int  | No       | Page number        |

**Authentication:** Required.

**Response:** `MetaResponse` — `{ page, total_pages, channels: ChannelSmall[] }`

---

## 13. Data Structures

### BigCoub

The primary coub object returned by single-coub and timeline endpoints.

| Field                       | Type                        | Notes                                              |
|-----------------------------|-----------------------------|----------------------------------------------------|
| `id`                        | int                         | Numeric coub ID                                    |
| `type`                      | string                      | `Coub::Simple` or `Coub::Recoub`                   |
| `title`                     | string                      |                                                    |
| `permalink`                 | string                      | Short identifier used in URLs                      |
| `visibility_type`           | string                      | `public`, `friends`, `unlisted`, `private`         |
| `channel_id`                | int                         |                                                    |
| `created_at`                | datetime (ISO 8601)         |                                                    |
| `updated_at`                | datetime (ISO 8601)         |                                                    |
| `is_done`                   | bool                        | Whether encoding is complete                       |
| `duration`                  | float                       | Loop duration in seconds                           |
| `views_count`               | int                         |                                                    |
| `likes_count`               | int                         |                                                    |
| `dislikes_count`            | int                         |                                                    |
| `recoubs_count`             | int                         |                                                    |
| `cotd`                      | bool or null                | Coub of the Day flag                               |
| `cotd_at`                   | date or null                |                                                    |
| `recoub`                    | bool or null                | Whether the authenticated user has recoubbed this  |
| `like`                      | bool or null                | Whether the authenticated user has liked this      |
| `recoub_to`                 | SubCoub or null             | Present if this coub is a recoub                   |
| `original_sound`            | bool                        |                                                    |
| `has_sound`                 | bool                        |                                                    |
| `file_versions`             | CoubFileVersion             | Contains all downloadable stream URLs              |
| `audio_versions`            | CoubAudioVersions or `{}`   |                                                    |
| `image_versions`            | CoubImageVersion or `{}`    | Thumbnail images                                   |
| `first_frame_versions`      | CoubFirstFrameVersion       | First-frame preview images                         |
| `dimensions`                | object                      | `{ "big": [w, h], "small": [w, h] }`              |
| `age_restricted`            | bool                        |                                                    |
| `allow_reuse`               | bool                        |                                                    |
| `banned`                    | bool                        |                                                    |
| `external_download`         | bool or CoubExternalDownload| Source video info if created from external content |
| `channel`                   | ChannelSmall                |                                                    |
| `percent_done`              | int                         | Encoding progress (0–100)                          |
| `tags`                      | CoubTag[]                   |                                                    |
| `media_blocks`              | CoubMediaBlocks             | Structured source media info                       |
| `raw_video_thumbnail_url`   | string                      |                                                    |
| `raw_video_title`           | string or null              |                                                    |
| `video_block_banned`        | bool                        |                                                    |
| `audio_copyright_claim`     | string or null              |                                                    |
| `categories`                | Category[]                  |                                                    |
| `communities`               | Community[]                 |                                                    |
| `normalize_change_allowed`  | bool                        |                                                    |
| `promoted_id`               | string or null              |                                                    |
| `visible_on_explore`        | bool                        |                                                    |
| `visible_on_explore_root`   | bool                        |                                                    |
| `abuses`                    | bool or null                |                                                    |
| `audio_file_url`            | URL or null                 |                                                    |
| `raw_video_id`              | int or string               |                                                    |
| `favourite`                 | bool or null                | Auth only: whether user has bookmarked this        |
| `flag`                      | bool or null                |                                                    |
| `published`                 | bool or null                | Auth only                                          |
| `published_at`              | datetime or null            | Auth only                                          |
| `is_editable`               | bool or null                | Auth only                                          |
| `age_restricted_by_admin`   | bool or null                | Auth only                                          |
| `global_safe`               | bool or null                | Auth only                                          |
| `page_w_h`                  | int[] or null               | Auth only                                          |
| `site_w_h`                  | int[] or null               | Auth only                                          |
| `site_w_h_small`            | int[] or null               | Auth only                                          |

---

### CoubFileVersion

The `file_versions` object nested inside a `BigCoub`. Contains the actual media
stream URLs needed for downloading.

```json
{
  "html5": {
    "video": {
      "med":    { "url": "https://...", "size": 123456 },
      "high":   { "url": "https://...", "size": 234567 },
      "higher": { "url": "https://...", "size": 345678 }
    },
    "audio": {
      "med":  { "url": "https://...", "size": 54321 },
      "high": { "url": "https://...", "size": 65432, "sample_duration": 30.0 }
    }
  },
  "mobile": {
    "video": "https://...",
    "audio": ["https://...aac_or_mp3", "https://...mp3"]
  },
  "share": {
    "default": "https://..."
  }
}
```

**Video quality tiers**

| Key      | Approximate resolution | Notes                              |
|----------|------------------------|------------------------------------|
| `med`    | ~360p                  | Same file as `mobile.video`        |
| `high`   | ~720p                  |                                    |
| `higher` | ~900p                  | Not always available               |

**Audio quality notes**

| Location        | Format        | Bitrate              |
|-----------------|---------------|----------------------|
| `html5.audio.med`  | MP3 CBR    | ~128 Kbps            |
| `html5.audio.high` | MP3 VBR    | ~160 Kbps            |
| `mobile.audio[0]`  | AAC or MP3 | ~128 Kbps CBR        |
| `mobile.audio[1]`  | MP3        | ~128 Kbps CBR        |

`size` may be `0` or `null` to indicate the stream is unavailable.  
`mobile.audio` is a raw URL string, not an object — no `size` field is provided.

**Share stream**

`file_versions.share.default` is a combined audio+video stream (approximately
720p, AAC audio). Shorter than the standalone audio tracks. AAC audio even when
the standalone mobile audio is MP3-only. May be `null`, `"{}"`, or an empty string
if not yet available.

---

### ChannelBig

| Field                    | Type            | Description                             |
|--------------------------|-----------------|-----------------------------------------|
| `id`                     | int             |                                         |
| `user_id`                | int             |                                         |
| `permalink`              | string          | URL slug                                |
| `title`                  | string          |                                         |
| `description`            | string or null  |                                         |
| `contacts`               | object or null  | Homepage, Tumblr, YouTube, Vimeo links  |
| `created_at`             | datetime        |                                         |
| `updated_at`             | datetime        |                                         |
| `avatar_versions`        | AvatarVersions  | `{ template, versions }`               |
| `followers_count`        | int             |                                         |
| `following_count`        | int             |                                         |
| `recoubs_count`          | int             |                                         |
| `likes_count`            | int             |                                         |
| `stories_count`          | int or null     |                                         |
| `views_count`            | int             |                                         |
| `hide_owner`             | bool            |                                         |
| `authentications`        | object[] or null| Linked social providers                 |
| `background_coub`        | BigCoub or null | Pinned background coub                  |
| `background_image`       | string or null  |                                         |
| `timeline_banner_image`  | string or null  |                                         |
| `meta`                   | object          | Social links: homepage, twitter, etc.   |
| `simple_coubs_count`     | int or null     |                                         |
| `i_follow_him`           | bool or null    | Auth only                               |
| `he_follows_me`          | bool or null    | Auth only                               |

---

### ChannelSmall

A compact channel object embedded inside coub and timeline responses.

| Field              | Type            | Description                 |
|--------------------|-----------------|-----------------------------|
| `id`               | int             |                             |
| `permalink`        | string          |                             |
| `title`            | string          |                             |
| `description`      | string or null  |                             |
| `followers_count`  | int             |                             |
| `following_count`  | int             |                             |
| `avatar_versions`  | AvatarVersions  | `{ template, versions }`   |
| `i_follow_him`     | bool or null    | Auth only                   |

---

### CoubTag

| Field   | Type   | Description                  |
|---------|--------|------------------------------|
| `id`    | int    |                              |
| `title` | string | Human-readable tag label     |
| `value` | string | URL-encoded tag value        |

---

## 14. Enumerated Values

### Category

Community/category permalink values used in timeline and channel endpoints.

| Value                 | Label                |
|-----------------------|----------------------|
| `animals-pets`        | Animals & Pets       |
| `mashup`              | Mashup               |
| `anime`               | Anime                |
| `movies`              | Movies               |
| `gaming`              | Gaming               |
| `cartoons`            | Cartoons             |
| `art`                 | Art                  |
| `music`               | Music                |
| `sports`              | Sports               |
| `science-technology`  | Science & Tech       |
| `celebrity`           | Celebrity            |
| `nature-travel`       | Nature & Travel      |
| `fashion`             | Fashion              |
| `cars`                | Cars                 |
| `nsfw`                | NSFW                 |
| `dance`               | Dance                |
| `news`                | News                 |
| `featured`            | Featured (special)   |
| `coub-of-the-day`     | Coub of the Day (special) |

---

### Visibility type

| Value       | Description                         |
|-------------|-------------------------------------|
| `public`    | Visible to everyone                 |
| `friends`   | Visible to followers only           |
| `unlisted`  | Accessible by direct link           |
| `private`   | Visible to owner only               |

---

### Social provider (Friends endpoint)

| Value        |
|--------------|
| `facebook`   |
| `twitter`    |
| `vkontakte`  |
| `google`     |

---

## 15. Notes and Limitations

### Page limits

| Timeline type         | Max pages |
|-----------------------|-----------|
| Tags                  | 99        |
| Hot section           | 99        |
| Communities           | 99        |
| Liked coubs           | 999       |
| Channel timeline      | No documented hard cap (soft limit around 999 observed) |

Requests beyond the tag/community/hot limit silently redirect back to page 1.
Requests beyond page 999 on other timelines return data for the last valid page
rather than an error.

### Rate limiting

No official rate limit documentation is published by Coub. In practice, rapid
automated requests may result in temporary `429 Too Many Requests` responses.
The downloader tools in the reference repositories add a configurable sleep
interval (`1` second default) between API calls as a courtesy.

### Media URL lifetime

Stream URLs returned by the API (`file_versions.*`) are served from a CDN and
are not permanent. URLs should be fetched fresh immediately before download
rather than stored for extended periods.

### Authentication token types

Two authentication mechanisms are observed:

1. **`access_token` query parameter** — An OAuth-style bearer token returned in
   `GET /api/v2/users/me` under the `api_token` field. Passed as a URL query
   parameter to authenticated endpoints.

2. **`remember_token` cookie** — A session cookie set by the coub.com web
   application after login. Passed as a `Cookie: remember_token=<value>` HTTP
   header. Used by browser-based session flows; also accepted by the API on
   endpoints that check authenticated user context.

### NSFW content

The `not_safe_for_work` field in raw coub JSON was announced for removal from
the API on 2022-06-27. Applications should not rely on it for content filtering.
The `age_restricted` and `age_restricted_by_admin` fields are the current
mechanisms.

### Coub upload workflow

The multi-step upload flow is:

```
POST /init_upload  →  POST /upload_video  →  POST /upload_audio  →  POST /finalize_upload
                                                                            ↓
                                              GET /finalize_status  (poll until done)
```

The `content_type` header for video upload must match the actual file format.
A reference gist of supported MIME types is available at:
`https://gist.github.com/Derfirm/5b11f77d64816153024e979141b69800`

### Known special-purpose channels

| Permalink      | Purpose                  |
|----------------|--------------------------|
| `royal.coubs`  | Editor's Choice coubs    |
| `oftheday`     | Coub of the Day coubs    |
