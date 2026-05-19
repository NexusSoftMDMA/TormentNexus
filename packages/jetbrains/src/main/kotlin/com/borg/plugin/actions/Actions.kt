package com.HyperCode.plugin.actions

import com.HyperCode.plugin.HypercodeService
========
package com.borg.plugin.actions

import com.borg.plugin.BorgService
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.ui.Messages

class ConnectAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = project.getService(HypercodeService::class.java)
        
        val url = Messages.showInputDialog(
            project,
            "Enter HyperCode Hub URL:",
            "Connect to HyperCode Hub",
=        val service = project.getService(BorgService::class.java)

        val url = Messages.showInputDialog(
            project,
            "Enter borg Hub URL:",
            "Connect to borg Hub",
>            null,
            "http://localhost:3000",
            null
        ) ?: return
        
        service.setHubUrl(url)
        if (service.connect()) {
            notify(project, "Connected to HyperCode Hub", NotificationType.INFORMATION)
        } else {
            notify(project, "Failed to connect to HyperCode Hub", NotificationType.ERROR)
========
            notify(project, "Connected to borg Hub", NotificationType.INFORMATION)
        } else {
            notify(project, "Failed to connect to borg Hub", NotificationType.ERROR)
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
        }
    }
    
    private fun notify(project: com.intellij.openapi.project.Project, message: String, type: NotificationType) {
        NotificationGroupManager.getInstance()
            .getNotificationGroup("HyperCode Notifications")
========
            .getNotificationGroup("borg Notifications")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
            .createNotification(message, type)
            .notify(project)
    }
}

class DisconnectAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = project.getService(HypercodeService::class.java)
        service.disconnect()
        
        NotificationGroupManager.getInstance()
            .getNotificationGroup("HyperCode Notifications")
            .createNotification("Disconnected from HyperCode Hub", NotificationType.INFORMATION)
========
        val service = project.getService(BorgService::class.java)
        service.disconnect()

        NotificationGroupManager.getInstance()
            .getNotificationGroup("borg Notifications")
            .createNotification("Disconnected from borg Hub", NotificationType.INFORMATION)
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
            .notify(project)
    }
}

class StartDebateAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val editor = e.getData(CommonDataKeys.EDITOR) ?: return
        val file = e.getData(CommonDataKeys.VIRTUAL_FILE) ?: return
        val service = project.getService(HypercodeService::class.java)
        
        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to HyperCode Hub", "HyperCode")
========
        val service = project.getService(BorgService::class.java)

        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to borg Hub", "borg")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
            return
        }
        
        val description = Messages.showInputDialog(
            project,
            "Describe what to debate:",
            "Start Council Debate",
            null
        ) ?: return
        
        val selection = editor.selectionModel
        val context = if (selection.hasSelection()) {
            selection.selectedText ?: ""
        } else {
            editor.document.text
        }
        
        val result = service.startDebate(description, file.path, context)
        
        if (result != null) {
            Messages.showInfoMessage(
                project,
                "Decision: ${result.decision}\nConsensus: ${result.consensusLevel}%\n\n${result.reasoning}",
                "Council Debate Result"
            )
        } else {
            Messages.showErrorDialog(project, "Debate failed", "HyperCode")
========
            Messages.showErrorDialog(project, "Debate failed", "borg")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
=======
class StartDebateAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val editor = e.getData(CommonDataKeys.EDITOR)
        val selectedText = editor?.selectionModel?.selectedText ?: ""

        val topic = Messages.showInputDialog(project, "Enter debate topic:", "Council Debate", null)
        if (topic != null) {
            val service = project.getService(HypercodeService::class.java)
========
            val service = project.getService(BorgService::class.java)
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
            service.startDebate(topic, selectedText) { result ->
                Messages.showInfoMessage(project, result, "Debate Result")
            }
>>>>>>> a3fab027fd172b66d6a0ec76e91f86354afa48e0
        }
    }
}

class ArchitectModeAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = project.getService(HypercodeService::class.java)
        
        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to HyperCode Hub", "HyperCode")
=        val service = project.getService(BorgService::class.java)

        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to borg Hub", "borg")
>            return
        }
        
        val task = Messages.showInputDialog(
            project,
            "Describe the task for reasoning:",
            "Architect Mode",
            null
        ) ?: return
        
        val session = service.startArchitectSession(task)
        
        if (session != null) {
            val approve = Messages.showYesNoDialog(
                project,
                "Session: ${session.sessionId}\nStatus: ${session.status}\n\n${session.plan?.description ?: "No plan yet"}\n\nApprove this plan?",
                "Architect Session",
                Messages.getQuestionIcon()
            )
            
            if (approve == Messages.YES) {
                service.approveArchitectPlan(session.sessionId)
                NotificationGroupManager.getInstance()
                    .getNotificationGroup("HyperCode Notifications")
========
                    .getNotificationGroup("borg Notifications")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
                    .createNotification("Plan approved", NotificationType.INFORMATION)
                    .notify(project)
            }
        } else {
            Messages.showErrorDialog(project, "Failed to start architect session", "HyperCode")
========
            Messages.showErrorDialog(project, "Failed to start architect session", "borg")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
        }
    }
}

class ViewAnalyticsAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = project.getService(HypercodeService::class.java)
        
        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to HyperCode Hub", "HyperCode")
========
        val service = project.getService(BorgService::class.java)

        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to borg Hub", "borg")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
            return
        }
        
        val summary = service.getAnalyticsSummary()
        
        if (summary != null) {
            Messages.showInfoMessage(
                project,
                """
                Total Supervisors: ${summary.totalSupervisors}
                Total Debates: ${summary.totalDebates}
                Approved: ${summary.totalApproved}
                Rejected: ${summary.totalRejected}
                Avg Consensus: ${summary.avgConsensus?.let { "%.1f%%".format(it) } ?: "N/A"}
                Avg Confidence: ${summary.avgConfidence?.let { "%.2f".format(it) } ?: "N/A"}
                """.trimIndent(),
                "Supervisor Analytics"
            )
        } else {
            Messages.showErrorDialog(project, "Failed to fetch analytics", "HyperCode")
========
            Messages.showErrorDialog(project, "Failed to fetch analytics", "borg")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
        }
    }
}

class RunAgentAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        Messages.showInfoMessage(project, "Run Agent feature coming soon", "HyperCode")
========
        Messages.showInfoMessage(project, "Run Agent feature coming soon", "borg")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
    }
}

class SearchMemoryAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        Messages.showInfoMessage(project, "Search Memory feature coming soon", "HyperCode")
========
        Messages.showInfoMessage(project, "Search Memory feature coming soon", "borg")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
    }
}
=======
        val task = Messages.showInputDialog(project, "Enter task for Architect:", "Architect Mode", null)
        if (task != null) {
            val service = project.getService(HypercodeService::class.java)
========
            val service = project.getService(BorgService::class.java)
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/borg/plugin/actions/Actions.kt
            service.startArchitectSession(task) { result ->
                Messages.showInfoMessage(project, result, "Architect Session Started")
            }
        }
    }
}
>>>>>>> a3fab027fd172b66d6a0ec76e91f86354afa48e0
