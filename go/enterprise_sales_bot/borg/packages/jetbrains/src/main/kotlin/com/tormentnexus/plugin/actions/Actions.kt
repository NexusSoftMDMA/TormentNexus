package com.TormentNexus.plugin.actions

import com.TormentNexus.plugin.TormentNexusService
========
package com.tormentnexus.plugin.actions

import com.tormentnexus.plugin.TormentNexusService
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.actionSystem.CommonDataKeys
import com.intellij.openapi.ui.Messages

class ConnectAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = project.getService(TormentNexusService::class.java)
        
        val url = Messages.showInputDialog(
            project,
            "Enter TormentNexus Hub URL:",
            "Connect to TormentNexus Hub",
=        val service = project.getService(TormentNexusService::class.java)

        val url = Messages.showInputDialog(
            project,
            "Enter tormentnexus Hub URL:",
            "Connect to tormentnexus Hub",
>            null,
            "http://localhost:3000",
            null
        ) ?: return
        
        service.setHubUrl(url)
        if (service.connect()) {
            notify(project, "Connected to TormentNexus Hub", NotificationType.INFORMATION)
        } else {
            notify(project, "Failed to connect to TormentNexus Hub", NotificationType.ERROR)
========
            notify(project, "Connected to tormentnexus Hub", NotificationType.INFORMATION)
        } else {
            notify(project, "Failed to connect to tormentnexus Hub", NotificationType.ERROR)
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
        }
    }
    
    private fun notify(project: com.intellij.openapi.project.Project, message: String, type: NotificationType) {
        NotificationGroupManager.getInstance()
            .getNotificationGroup("TormentNexus Notifications")
========
            .getNotificationGroup("tormentnexus Notifications")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
            .createNotification(message, type)
            .notify(project)
    }
}

class DisconnectAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = project.getService(TormentNexusService::class.java)
        service.disconnect()
        
        NotificationGroupManager.getInstance()
            .getNotificationGroup("TormentNexus Notifications")
            .createNotification("Disconnected from TormentNexus Hub", NotificationType.INFORMATION)
========
        val service = project.getService(TormentNexusService::class.java)
        service.disconnect()

        NotificationGroupManager.getInstance()
            .getNotificationGroup("tormentnexus Notifications")
            .createNotification("Disconnected from tormentnexus Hub", NotificationType.INFORMATION)
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
            .notify(project)
    }
}

class StartDebateAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val editor = e.getData(CommonDataKeys.EDITOR) ?: return
        val file = e.getData(CommonDataKeys.VIRTUAL_FILE) ?: return
        val service = project.getService(TormentNexusService::class.java)
        
        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to TormentNexus Hub", "TormentNexus")
========
        val service = project.getService(TormentNexusService::class.java)

        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to tormentnexus Hub", "tormentnexus")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
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
            Messages.showErrorDialog(project, "Debate failed", "TormentNexus")
========
            Messages.showErrorDialog(project, "Debate failed", "tormentnexus")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
=======
class StartDebateAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val editor = e.getData(CommonDataKeys.EDITOR)
        val selectedText = editor?.selectionModel?.selectedText ?: ""

        val topic = Messages.showInputDialog(project, "Enter debate topic:", "Council Debate", null)
        if (topic != null) {
            val service = project.getService(TormentNexusService::class.java)
========
            val service = project.getService(TormentNexusService::class.java)
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
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
        val service = project.getService(TormentNexusService::class.java)
        
        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to TormentNexus Hub", "TormentNexus")
=        val service = project.getService(TormentNexusService::class.java)

        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to tormentnexus Hub", "tormentnexus")
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
                    .getNotificationGroup("TormentNexus Notifications")
========
                    .getNotificationGroup("tormentnexus Notifications")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
                    .createNotification("Plan approved", NotificationType.INFORMATION)
                    .notify(project)
            }
        } else {
            Messages.showErrorDialog(project, "Failed to start architect session", "TormentNexus")
========
            Messages.showErrorDialog(project, "Failed to start architect session", "tormentnexus")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
        }
    }
}

class ViewAnalyticsAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        val service = project.getService(TormentNexusService::class.java)
        
        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to TormentNexus Hub", "TormentNexus")
========
        val service = project.getService(TormentNexusService::class.java)

        if (!service.isConnected()) {
            Messages.showErrorDialog(project, "Not connected to tormentnexus Hub", "tormentnexus")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
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
            Messages.showErrorDialog(project, "Failed to fetch analytics", "TormentNexus")
========
            Messages.showErrorDialog(project, "Failed to fetch analytics", "tormentnexus")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
        }
    }
}

class RunAgentAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        Messages.showInfoMessage(project, "Run Agent feature coming soon", "TormentNexus")
========
        Messages.showInfoMessage(project, "Run Agent feature coming soon", "tormentnexus")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
    }
}

class SearchMemoryAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        Messages.showInfoMessage(project, "Search Memory feature coming soon", "TormentNexus")
========
        Messages.showInfoMessage(project, "Search Memory feature coming soon", "tormentnexus")
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
    }
}
=======
        val task = Messages.showInputDialog(project, "Enter task for Architect:", "Architect Mode", null)
        if (task != null) {
            val service = project.getService(TormentNexusService::class.java)
========
            val service = project.getService(TormentNexusService::class.java)
>>>>>>>> origin/dependabot/cargo/packages/zed-extension/cargo-64b2a50fd2:packages/jetbrains/src/main/kotlin/com/tormentnexus/plugin/actions/Actions.kt
            service.startArchitectSession(task) { result ->
                Messages.showInfoMessage(project, result, "Architect Session Started")
            }
        }
    }
}
>>>>>>> a3fab027fd172b66d6a0ec76e91f86354afa48e0
