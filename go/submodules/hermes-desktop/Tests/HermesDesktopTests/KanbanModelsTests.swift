import Foundation
import Testing
@testable import HermesDesktop

struct KanbanModelsTests {
    @Test
    func boardPayloadDecodesRepresentativeUpstreamFields() throws {
        let data = """
        {
          "ok": true,
          "board": {
            "database_path": "~/.hermes/kanban.db",
            "host_wide": true,
            "is_initialized": true,
            "has_kanban_module": true,
            "has_hermes_cli": true,
            "dispatcher": {
              "running": false,
              "message": "No gateway is running"
            },
            "latest_event_id": 42,
            "tasks": [
              {
                "id": "t_1234abcd",
                "title": "Release candidate smoke test",
                "body": "Verify packaging, signing, and docs.",
                "assignee": "release",
                "status": "ready",
                "priority": 3,
                "created_by": "desktop",
                "created_at": 1800000000,
                "started_at": null,
                "completed_at": null,
                "workspace_kind": "scratch",
                "workspace_path": null,
                "tenant": "desktop",
                "result": null,
                "skills": ["release-check"],
                "spawn_failures": 1,
                "worker_pid": null,
                "last_spawn_error": "missing keychain",
                "max_runtime_seconds": 3600,
                "max_retries": 3,
                "last_heartbeat_at": null,
                "current_run_id": null,
                "parent_ids": ["t_parent"],
                "child_ids": ["t_child_a", "t_child_b", "t_child_c"],
                "progress": {
                  "done": 1,
                  "total": 3
                },
                "comment_count": 2,
                "event_count": 5,
                "run_count": 1,
                "latest_event_at": 1800000300,
                "warnings": {
                  "count": 2,
                  "kinds": {
                    "completion_blocked_hallucination": 1,
                    "suspected_hallucinated_references": 1
                  },
                  "latest_at": 1800000250
                }
              }
            ],
            "assignees": [
              {
                "name": "release",
                "on_disk": true,
                "counts": {
                  "ready": 1
                }
              }
            ],
            "tenants": ["desktop"],
            "stats": {
              "by_status": {
                "ready": 1
              },
              "by_assignee": {
                "release": {
                  "ready": 1
                }
              },
              "oldest_ready_age_seconds": 120,
              "now": 1800000400
            }
          }
        }
        """.data(using: .utf8)!

        let board = try JSONDecoder().decode(KanbanBoardResponse.self, from: data).board

        #expect(board.databasePath == "~/.hermes/kanban.db")
        #expect(board.hostWide)
        #expect(board.dispatcher?.isKnownInactive == true)
        #expect(board.tasks.count == 1)
        #expect(board.tasks[0].status == .ready)
        #expect(board.tasks[0].workspaceKind == .scratch)
        #expect(board.tasks[0].priorityLabel == "P+3")
        #expect(board.tasks[0].maxRetries == 3)
        #expect(board.tasks[0].progressLabel == "1/3 done")
        #expect(board.tasks[0].hasActiveWarnings)
        #expect(board.tasks[0].warnings?.includesBlockedCompletion == true)
        #expect(board.tasks[0].matchesSearch("reference warnings"))
        #expect(board.stats?.now == 1800000400)
        #expect(board.tasks[0].matchesSearch("packaging"))
        #expect(board.tasks[0].matchesSearch("release-check"))
        #expect(board.tasks[0].matchesSearch("t_parent"))
        #expect(!board.tasks[0].matchesSearch("unrelated"))
    }

    @Test
    func taskDetailPayloadDecodesCommentsEventsRunsAndWorkerLog() throws {
        let data = """
        {
          "ok": true,
          "detail": {
            "task": {
              "id": "t_done",
              "title": "Ship notes",
              "body": "",
              "assignee": null,
              "status": "done",
              "priority": 0,
              "created_by": "desktop",
              "created_at": 1800000000,
              "started_at": 1800000100,
              "completed_at": 1800000200,
              "workspace_kind": "dir",
              "workspace_path": "/tmp/release",
              "tenant": null,
              "result": "Published.",
              "skills": [],
              "spawn_failures": 0,
              "worker_pid": null,
              "last_spawn_error": null,
              "max_runtime_seconds": null,
              "max_retries": null,
              "last_heartbeat_at": null,
              "current_run_id": null,
              "parent_ids": [],
              "child_ids": ["t_child"],
              "progress": null,
              "comment_count": 1,
              "event_count": 1,
              "run_count": 1,
              "latest_event_at": 1800000200
            },
            "parent_ids": [],
            "child_ids": ["t_child"],
            "comments": [
              {
                "id": 7,
                "task_id": "t_done",
                "author": "desktop",
                "body": "Looks good.",
                "created_at": 1800000150
              }
            ],
            "events": [
              {
                "id": 8,
                "task_id": "t_done",
                "kind": "completed",
                "payload": {
                  "summary": "Published."
                },
                "created_at": 1800000200,
                "run_id": 2
              }
            ],
            "runs": [
              {
                "id": 2,
                "task_id": "t_done",
                "profile": "release",
                "step_key": null,
                "status": "done",
                "outcome": "completed",
                "summary": "Published.",
                "error": null,
                "metadata": {
                  "changed_files": 3
                },
                "worker_pid": 123,
                "started_at": 1800000100,
                "ended_at": 1800000200
              }
            ],
            "worker_log": "done"
          }
        }
        """.data(using: .utf8)!

        let detail = try JSONDecoder().decode(KanbanTaskDetailResponse.self, from: data).detail

        #expect(detail.task.status == .done)
        #expect(detail.task.workspaceKind == .directory)
        #expect(detail.task.trimmedResult == "Published.")
        #expect(detail.comments[0].body == "Looks good.")
        #expect(detail.events[0].payload?["summary"] == .string("Published."))
        #expect(detail.runs[0].resolvedOutcome == "completed")
        #expect(detail.workerLog == "done")
        #expect(detail.homeChannels.isEmpty)
    }

    @Test
    func boardsPayloadDecodesProjectsAndCurrentSelection() throws {
        let data = """
        {
          "ok": true,
          "current": "desktop",
          "supports_board_management": true,
          "boards": [
            {
              "slug": "default",
              "name": "Default",
              "description": "",
              "icon": "",
              "color": "",
              "created_at": null,
              "archived": false,
              "db_path": "~/.hermes/kanban.db",
              "is_current": false,
              "counts": {
                "ready": 2
              },
              "total": 2
            },
            {
              "slug": "desktop",
              "name": "Hermes Desktop",
              "description": "Release work",
              "icon": "desktopcomputer",
              "color": "#4F8CFF",
              "created_at": 1800000000,
              "archived": false,
              "db_path": "~/.hermes/kanban/boards/desktop/kanban.db",
              "is_current": true,
              "counts": {
                "ready": 1,
                "done": 3
              },
              "total": 4
            }
          ]
        }
        """.data(using: .utf8)!

        let response = try JSONDecoder().decode(KanbanBoardsResponse.self, from: data)

        #expect(response.current == "desktop")
        #expect(response.supportsBoardManagement)
        #expect(response.boards.count == 2)
        #expect(response.boards[0].isDefault)
        #expect(response.boards[1].resolvedName == "Hermes Desktop")
        #expect(response.boards[1].taskTotal == 4)
        #expect(response.boards[1].databasePath == "~/.hermes/kanban/boards/desktop/kanban.db")
        #expect(response.boards[1].isCurrent)
    }

    @Test
    func taskDetailPayloadDecodesHomeChannels() throws {
        let task = makeTask(status: .ready)
        let payload = TaskDetailFixture(
            task: task,
            homeChannels: [
                KanbanHomeChannel(
                    platform: "telegram",
                    chatID: "123",
                    threadID: "",
                    name: "Home",
                    subscribed: true
                )
            ]
        )
        let data = try JSONEncoder().encode(KanbanTaskDetailResponse(ok: true, detail: payload.detail))
        let detail = try JSONDecoder().decode(KanbanTaskDetailResponse.self, from: data).detail

        #expect(detail.homeChannels.count == 1)
        #expect(detail.homeChannels[0].platform == "telegram")
        #expect(detail.homeChannels[0].subscribed)
        #expect(detail.homeChannels[0].destinationLabel == "123")
    }

    @Test
    func draftValidationRequiresTitleAndNormalizesLists() {
        var draft = KanbanTaskDraft()
        #expect(draft.validationError == "Task title is required.")

        draft.title = "  Prepare changelog  "
        draft.body = "  Include Kanban work  "
        draft.assignee = " release "
        draft.tenant = " desktop "
        draft.priority = 42
        draft.maxRetriesText = " 3 "
        draft.skillsText = "release-notes, docs, "
        draft.parentIDsText = " t_parent_a, t_parent_b\n t_parent_a "

        #expect(draft.validationError == nil)
        #expect(draft.normalizedTitle == "Prepare changelog")
        #expect(draft.normalizedBody == "Include Kanban work")
        #expect(draft.normalizedAssignee == "release")
        #expect(draft.normalizedTenant == "desktop")
        #expect(draft.priority == 42)
        #expect(draft.normalizedMaxRetries == 3)
        #expect(draft.skills == ["release-notes", "docs"])
        #expect(draft.parentIDs == ["t_parent_a", "t_parent_b"])

        draft.maxRetriesText = "0"
        #expect(draft.validationError == "Max retries must be a whole number greater than 0.")
    }

    @Test
    func boardDraftValidationMatchesUpstreamSlugRules() {
        var draft = KanbanBoardDraft()
        #expect(draft.validationError == "Board slug is required.")

        draft.slug = "  Desktop_Release-1  "
        draft.name = " Hermes Desktop "
        draft.description = " Release stream "
        draft.icon = " desktopcomputer "
        draft.color = " #4F8CFF "

        #expect(draft.validationError == nil)
        #expect(draft.normalizedSlug == "desktop_release-1")
        #expect(draft.normalizedName == "Hermes Desktop")
        #expect(draft.normalizedDescription == "Release stream")
        #expect(draft.normalizedIcon == "desktopcomputer")
        #expect(draft.normalizedColor == "#4F8CFF")

        draft.slug = "_bad"
        #expect(draft.validationError == "Board slug must be 1-64 lowercase letters, numbers, hyphens, or underscores.")
    }

    @Test
    func taskActionsMatchUpstreamKernelTransitions() throws {
        let ready = makeTask(status: .ready)
        #expect(ready.canBlock)
        #expect(ready.canComplete)
        #expect(!ready.canUnblock)
        #expect(!ready.canSpecify)

        let todo = makeTask(status: .todo)
        #expect(!todo.canBlock)
        #expect(!todo.canComplete)
        #expect(!todo.canUnblock)
        #expect(!todo.canSpecify)

        let triage = makeTask(status: .triage)
        #expect(triage.canSpecify)

        let blocked = makeTask(status: .blocked)
        #expect(!blocked.canBlock)
        #expect(blocked.canComplete)
        #expect(blocked.canUnblock)
    }

    private func makeTask(status: KanbanTaskStatus) -> KanbanTask {
        KanbanTask(
            id: "t_test",
            title: "Test",
            body: nil,
            assignee: nil,
            status: status,
            priority: 0,
            createdBy: nil,
            createdAt: nil,
            startedAt: nil,
            completedAt: nil,
            workspaceKind: .scratch,
            workspacePath: nil,
            tenant: nil,
            result: nil,
            skills: [],
            spawnFailures: 0,
            workerPID: nil,
            lastSpawnError: nil,
            maxRuntimeSeconds: nil,
            maxRetries: nil,
            lastHeartbeatAt: nil,
            currentRunID: nil,
            parentIDs: [],
            childIDs: [],
            progress: nil,
            commentCount: 0,
            eventCount: 0,
            runCount: 0,
            latestEventAt: nil
        )
    }

    private struct TaskDetailFixture: Encodable {
        let detail: KanbanTaskDetail

        init(task: KanbanTask, homeChannels: [KanbanHomeChannel]) {
            detail = KanbanTaskDetail(
                task: task,
                parentIDs: [],
                childIDs: [],
                comments: [],
                events: [],
                runs: [],
                workerLog: nil,
                homeChannels: homeChannels
            )
        }
    }
}
