# Slack Explorer MCP E2E Test Cases

E2E test cases to manually verify Slack Explorer MCP functionality on Claude Code after making changes.

## Overview

- **Purpose**: Verify that each MCP tool works correctly with the actual Slack API
- **Execution**: Run each test case on Claude Code and check the response

## Test Strategy

- **Normal cases**: Try with default sample values first, adjust only if no results are returned
- **Error cases**: Use fixed invalid IDs (e.g., `U000000000`)
- **Exploratory approach**: Use results from previous tools to test subsequent tools (e.g., search_messages → get_thread_replies)
- **Parallel execution**: Run independent test cases in parallel for efficiency
  - Error cases for each tool can be run in parallel since they have no dependencies
  - Exploratory tests (those using results from previous tools) must run sequentially

---

## Test Cases

### search_messages

#### Basic query search

**Steps:**
1. Search for messages with "hello"
2. If no results, retry with a keyword used in your workspace

**Success Criteria:**
- [ ] Response is valid JSON
- [ ] `workspace_url` field exists
- [ ] `messages.matches` array is returned
- [ ] Each message has `user`, `text`, `ts`, `channel` fields
- [ ] `messages.pagination` contains pagination info

---

#### Search with filters

**Steps:**
1. Search for messages in "random" channel before 2025-01-01
2. If channel doesn't exist, change to a channel name in your workspace and retry

**Success Criteria:**
- [ ] Response is valid JSON
- [ ] All returned messages are from the specified channel
- [ ] `in_channel` and `before` filters are correctly applied

---

#### Error when query contains modifiers

**Steps:**
1. Search using search_messages tool with query "from:someone test"

**Expected Response:**
- Error message: "query field cannot contain modifiers (from:, in:, etc.). Please use the dedicated fields"

**Success Criteria:**
- [ ] Error message is returned
- [ ] Message explains that modifiers should not be in the query field

---

### get_thread_replies

#### Get thread replies

**Steps:**
1. Search for messages with "meeting" using search_messages
2. Find a message with `thread_ts`
   - If not found, retry with different keywords ("question", "help", etc.)
3. Use the `channel.id` and `thread_ts` from the found message to call get_thread_replies

**Success Criteria:**
- [ ] Response is valid JSON
- [ ] `messages` array contains parent message and replies
- [ ] Each message has `user`, `text`, `ts` fields
- [ ] `has_more` field exists

---

#### Error with non-existent thread_ts

**Steps:**
1. Call get_thread_replies with channel_id "C000000000" and thread_ts "0000000000.000000"

**Success Criteria:**
- [ ] Error response is returned
- [ ] Error indicates channel or thread not found

---

### get_user_profiles

#### Get multiple user profiles (normal and error mixed)

**Steps:**
1. Search for users with "a" using partial match in search_users_by_name to find 2+ users
   - If not found, retry with a different character
2. Call get_user_profiles with the 2 found user_ids plus non-existent user_id "UERROR00000"

**Success Criteria:**
- [ ] Response is valid JSON array
- [ ] Existing user profiles have `user_id`, `display_name`, `real_name` fields
- [ ] Non-existent user (UERROR00000) entry has `error` field

---

### search_users_by_name

#### Exact match search

**Steps:**
1. Search for users with "a" using partial match in search_users_by_name to find a user
2. Use that user's `display_name` to perform exact match search (exact=true)

**Success Criteria:**
- [ ] Response is valid JSON array
- [ ] Matching user is returned
- [ ] Display name exactly matches the search term

---

#### Partial match search (exact=false)

**Steps:**
1. Search for "a" with partial match (exact=false)

**Success Criteria:**
- [ ] Response is valid JSON array
- [ ] Users with partial matches are returned
- [ ] Each user has `user_id`, `display_name`, `real_name` fields

---

#### Search for non-existent name

**Steps:**
1. Search for exact match with display name "ThisUserDoesNotExist12345"

**Success Criteria:**
- [ ] Response is valid JSON array
- [ ] Array is empty `[]`

---

### search_files

#### Search with query

**Steps:**
1. Search for files with "report"
2. If no results, retry with different keywords ("doc", "image", etc.)

**Success Criteria:**
- [ ] Response is valid JSON
- [ ] `files` array is returned
- [ ] Each file has `id`, `title`, `filetype`, `permalink` fields
- [ ] `pagination` contains pagination info

---

#### Search by file type

**Steps:**
1. Search for files with types "canvases"
2. If no canvases found, retry with types "images" or "pdfs"

**Success Criteria:**
- [ ] Response is valid JSON
- [ ] `files` array is returned
- [ ] `types` filter is correctly applied

---

#### Error when query contains modifiers

**Steps:**
1. Search using search_files tool with query "type:pdf report"

**Expected Response:**
- Error message: "query field cannot contain modifiers (from:, in:, type:, etc.). Please use the dedicated fields"

**Success Criteria:**
- [ ] Error message is returned
- [ ] Message explains that modifiers should not be in the query field

---

### get_canvas_content

#### Get canvas content

**Steps:**
1. Search for canvases using search_files with types "canvases"
2. Use the found canvas `id` to call get_canvas_content
   - Skip this test if no canvases are found

**Success Criteria:**
- [ ] Response is valid JSON
- [ ] `canvases` array is returned
- [ ] Canvas entry has `id`, `title`, `content`, `permalink` fields
- [ ] `content` contains HTML content

---

#### Error with non-existent canvas ID

**Steps:**
1. Call get_canvas_content with canvas_ids "F000000000"

**Success Criteria:**
- [ ] Response is valid JSON
- [ ] Canvas entry has `error` field
- [ ] Error message indicates file not found

---

#### Error with invalid canvas ID format

**Steps:**
1. Call get_canvas_content with canvas_ids "INVALID_ID"

**Success Criteria:**
- [ ] Response is valid JSON
- [ ] Canvas entry has `error` field
- [ ] Error message indicates invalid ID format

---

## Test Case Summary

| Tool | Type | Description |
|------|------|-------------|
| search_messages | Normal | Basic query search |
| search_messages | Normal | Search with filters |
| search_messages | Error | Error when query contains modifiers |
| get_thread_replies | Normal | Get thread replies |
| get_thread_replies | Error | Error with non-existent thread_ts |
| get_user_profiles | Normal/Error | Get multiple user profiles (mixed) |
| search_users_by_name | Normal | Exact match search |
| search_users_by_name | Normal | Partial match search |
| search_users_by_name | Normal | Search for non-existent name (empty array) |
| search_files | Normal | Search with query |
| search_files | Normal | Search by file type |
| search_files | Error | Error when query contains modifiers |
| get_canvas_content | Normal | Get canvas content |
| get_canvas_content | Error | Error with non-existent canvas ID |
| get_canvas_content | Error | Error with invalid canvas ID format |
