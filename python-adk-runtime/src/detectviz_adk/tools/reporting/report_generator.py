from detectviz_adk.tools.remote_tool import RemoteTool

# A canonical instance of the ReportGenerator remote tool.
# Agents should import and use this object to interact with the Go plugin.
ReportGenerator = RemoteTool(
    tool_id="reporting.report_generator",
    tool_version="0.1.0"
)
