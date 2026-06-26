import Foundation
import Testing
@testable import HermesDesktop

struct CronJobModelsTests {
    @Test
    func scheduleFormatterCoversReleaseCriticalExpressions() {
        #expect(CronScheduleFormatter.humanReadableDescription(for: "30 9 * * 1-5") == "Every weekday at 09:30")
        #expect(CronScheduleFormatter.humanReadableDescription(for: "0 18 * * mon,wed,fri") == "Every Mon, Wed, Fri at 18:00")
        #expect(CronScheduleFormatter.humanReadableDescription(for: "15 * * * *") == "Every hour at :15")
        #expect(CronScheduleFormatter.humanReadableDescription(for: "every 3h") == "Every 3 hours")
        #expect(CronScheduleFormatter.humanReadableDescription(for: "10m") == "Once in 10 minutes")
    }

    @Test
    func cronJobSearchUsesCanonicalFieldsVisibleInTheApp() {
        let job = makeJob(
            name: "Morning Briefing",
            prompt: "Review overnight errors and ship a concise summary",
            skills: ["deploy-check", "triage"],
            model: "gpt-5",
            schedule: CronSchedule(kind: "cron", expr: "30 9 * * 1-5", timezone: "Europe/Rome"),
            scheduleDisplay: "",
            deliveryTarget: "Local Only"
        )

        #expect(job.rawScheduleText == "30 9 * * 1-5")
        #expect(job.resolvedScheduleDisplay == "Every weekday at 09:30")
        #expect(job.matchesSearch("briefing"))
        #expect(job.matchesSearch("GPT-5"))
        #expect(job.matchesSearch("deploy"))
        #expect(job.matchesSearch("weekday"))
        #expect(job.matchesSearch("local only"))
        #expect(!job.matchesSearch("nightly backup"))
    }

    @Test
    func scriptOnlyJobsDecodeAndValidateWithoutPrompt() throws {
        let data = """
        {
          "id": "job-watch",
          "name": "Memory watchdog",
          "prompt": "",
          "script": "memory-watchdog.sh",
          "workdir": "/srv/hermes",
          "no_agent": true,
          "skills": [],
          "model": null,
          "provider": null,
          "base_url": null,
          "schedule": {
            "kind": "cron",
            "expr": "*/5 * * * *",
            "timezone": "UTC"
          },
          "schedule_display": "*/5 * * * *",
          "recurrence": null,
          "enabled": true,
          "state": "scheduled",
          "created_at": null,
          "next_run_at": null,
          "last_run_at": null,
          "last_status": null,
          "last_error": null,
          "delivery_target": "telegram",
          "origin": null,
          "last_delivery_error": null
        }
        """.data(using: .utf8)!

        let job = try JSONDecoder().decode(CronJob.self, from: data)

        #expect(job.noAgent)
        #expect(job.trimmedScript == "memory-watchdog.sh")
        #expect(job.trimmedWorkdir == "/srv/hermes")
        #expect(job.previewPrompt == "Script-only watchdog: memory-watchdog.sh")
        #expect(job.matchesSearch("script only"))

        var draft = CronJobDraft(job: job)
        #expect(draft.validationError == nil)
        draft.script = ""
        #expect(draft.validationError == "A script path is required for script-only jobs.")
    }

    @Test
    func cronStatePresentationFallsBackSafely() {
        #expect(CronJobState.scheduled.displayTitle(isEnabled: true) == "Active")
        #expect(CronJobState.paused.displayTitle(isEnabled: true) == "Paused")
        #expect(CronJobState.running.displayTitle(isEnabled: true) == "Running")
        #expect(CronJobState.other("").displayTitle(isEnabled: true) == "Active")
        #expect(CronJobState.other("").displayTitle(isEnabled: false) == "Paused")
        #expect(CronJobState.other("waiting_for_slot").displayTitle(isEnabled: true) == "Waiting For Slot")
        #expect(CronJobState.running.isActive)
        #expect(!CronJobState.failed.isActive)
    }

    @Test
    func deliveryPresetRecognizesAllChannelsValue() {
        #expect(CronDeliveryPreset.from(deliveryTarget: "all").preset == .all)

        var draft = CronJobDraft()
        draft.deliveryPreset = .all
        #expect(draft.normalizedDeliveryTarget == "all")
    }

    private func makeJob(
        name: String,
        prompt: String,
        skills: [String],
        model: String?,
        schedule: CronSchedule?,
        scheduleDisplay: String,
        deliveryTarget: String
    ) -> CronJob {
        CronJob(
            id: "job-1",
            name: name,
            prompt: prompt,
            skills: skills,
            model: model,
            provider: "openai",
            baseURL: nil,
            schedule: schedule,
            scheduleDisplay: scheduleDisplay,
            recurrence: nil,
            enabled: true,
            state: .scheduled,
            createdAt: nil,
            nextRunAt: nil,
            lastRunAt: nil,
            lastStatus: nil,
            lastError: nil,
            deliveryTarget: deliveryTarget,
            origin: nil,
            lastDeliveryError: nil
        )
    }
}
