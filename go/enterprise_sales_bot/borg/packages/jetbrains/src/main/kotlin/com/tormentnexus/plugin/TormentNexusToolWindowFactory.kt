package com.tormentnexus.plugin

import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.ToolWindow
import com.intellij.openapi.wm.ToolWindowFactory
import com.intellij.ui.components.JBLabel
import com.intellij.ui.components.JBScrollPane
import com.intellij.ui.content.ContentFactory
import java.awt.BorderLayout
import javax.swing.*

class TormentNexusToolWindowFactory : ToolWindowFactory {
    
    override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
        val panel = TormentNexusToolWindowPanel(project)
    override fun createToolWindowContent(project: Project, toolWindow: ToolWindow) {
        val panel = JPanel(BorderLayout())
        val textArea = JTextArea("tormentnexus Hub Status: Connected\n\nWaiting for activity...")
        textArea.isEditable = false
        
        val scrollPane = JScrollPane(textArea)
        panel.add(scrollPane, BorderLayout.CENTER)
        
        val bottomPanel = JPanel()
        val refreshButton = JButton("Refresh")
        refreshButton.addActionListener {
            textArea.append("\nRefreshing state...")
        }
        bottomPanel.add(refreshButton)
        panel.add(bottomPanel, BorderLayout.SOUTH)

        val content = ContentFactory.getInstance().createContent(panel, "", false)
        toolWindow.contentManager.addContent(content)
    }
}

class TormentNexusToolWindowPanel(private val project: Project) : JPanel(BorderLayout()) {
    
    private val service = project.getService(TormentNexusService::class.java)
    private val statusLabel = JBLabel("Disconnected")
    private val outputArea = JTextArea().apply {
        isEditable = false
        lineWrap = true
        wrapStyleWord = true
    }
    
    init {
        val topPanel = JPanel().apply {
            layout = BoxLayout(this, BoxLayout.X_AXIS)
            add(JBLabel("tormentnexus Hub: "))
            add(statusLabel)
            add(Box.createHorizontalGlue())
            add(JButton("Connect").apply {
                addActionListener { connect() }
            })
            add(JButton("Refresh").apply {
                addActionListener { refreshAnalytics() }
            })
        }
        
        add(topPanel, BorderLayout.NORTH)
        add(JBScrollPane(outputArea), BorderLayout.CENTER)
        
        val buttonPanel = JPanel().apply {
            layout = BoxLayout(this, BoxLayout.Y_AXIS)
            add(JButton("View Analytics").apply {
                addActionListener { refreshAnalytics() }
            })
            add(JButton("List Templates").apply {
                addActionListener { listTemplates() }
            })
        }
        add(buttonPanel, BorderLayout.EAST)
    }
    
    private fun connect() {
        if (service.connect()) {
            statusLabel.text = "Connected"
            appendOutput("Connected to tormentnexus Hub")
            refreshAnalytics()
        } else {
            statusLabel.text = "Connection Failed"
            appendOutput("Failed to connect to tormentnexus Hub")
        }
    }
    
    private fun refreshAnalytics() {
        val summary = service.getAnalyticsSummary()
        if (summary != null) {
            appendOutput("""
                === Supervisor Analytics ===
                Total Supervisors: ${summary.totalSupervisors}
                Total Debates: ${summary.totalDebates}
                Approved: ${summary.totalApproved}
                Rejected: ${summary.totalRejected}
                Avg Consensus: ${summary.avgConsensus?.let { "%.1f%%".format(it) } ?: "N/A"}
                Avg Confidence: ${summary.avgConfidence?.let { "%.2f".format(it) } ?: "N/A"}
            """.trimIndent())
        } else {
            appendOutput("Failed to fetch analytics")
        }
    }
    
    private fun listTemplates() {
        val templates = service.getDebateTemplates()
        if (templates.isNotEmpty()) {
            appendOutput("\n=== Debate Templates ===")
            templates.forEach { t ->
                appendOutput("â€¢ ${t.name} (${t.id}): ${t.description ?: "No description"}")
            }
        } else {
            appendOutput("No templates available or failed to fetch")
        }
    }
    
    private fun appendOutput(text: String) {
        outputArea.append("$text\n\n")
        outputArea.caretPosition = outputArea.document.length
    }
}
