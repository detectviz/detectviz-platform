# {{ t('incident_report') }}

**{{ t('generated_at') }}**: {{ report_meta.generated_at | format_timestamp(language) }}  
**{{ t('report_version') }}**: {{ report_meta.version }}  
**{{ t('language') }}**: {{ language }}

---

## {{ t('executive_summary') }}

{{ summary | default('事件摘要待補充') }}

## {{ t('incident_overview') }}

| 項目 | 詳情 |
|------|------|
| **{{ t('severity') }}** | {{ incident.severity | format_severity(language) | default('未知') }} |
| **{{ t('affected_services') }}** | {% for service in incident.affected_services | default([]) %}{{ service }}{% if not loop.last %}, {% endif %}{% endfor %} |
| **{{ t('duration') }}** | {{ incident.duration | format_duration(language) | default('未知') }} |
| **{{ t('detection_time') }}** | {{ incident.detection_time | format_timestamp(language) | default('未知') }} |
| **{{ t('resolution_time') }}** | {{ incident.resolution_time | format_timestamp(language) | default('未知') }} |
| **{{ t('responsible_team') }}** | {{ incident.responsible_team | default('未指定') }} |

## {{ t('impact_analysis') }}

### {{ t('customer_impact') }}
{% if impact.customer_affected %}
- **影響客戶數**: {{ impact.customer_affected | default('未知') }}
- **影響地區**: {{ impact.affected_regions | join(', ') | default('未知') }}
{% endif %}

{{ impact.customer | default('客戶影響分析待補充') }}

### {{ t('business_impact') }}
{% if impact.revenue_loss %}
- **預估收入損失**: ${{ impact.revenue_loss | default(0) }}
{% endif %}

{{ impact.business | default('業務影響分析待補充') }}

### {{ t('technical_impact') }}
{% if impact.system_degradation %}
- **系統降級程度**: {{ impact.system_degradation }}%
- **資料丟失**: {% if impact.data_loss %}是{% else %}否{% endif %}
{% endif %}

{{ impact.technical | default('技術影響分析待補充') }}

## {{ t('system_metrics') }}

{% for chart in chart_embeddings | default([]) %}
{{ chart }}

{% endfor %}

{% if metrics %}
### 關鍵指標

| 指標 | 事件前 | 事件中 | 事件後 |
|------|--------|--------|--------|
{% for metric in metrics %}
| {{ metric.name }} | {{ metric.before }} | {{ metric.during }} | {{ metric.after }} |
{% endfor %}
{% endif %}

## {{ t('incident_timeline') }}

{% for event in timeline | default([]) %}
### {{ event.timestamp | format_timestamp(language) }}

**{{ event.title | default('事件') }}** {% if event.severity %}*[{{ event.severity | format_severity(language) }}]* {% endif %}

{{ event.description | default('無描述') }}

{% if event.actions %}
**採取的行動**:
{% for action in event.actions %}
- {{ action }}
{% endfor %}
{% endif %}

{% if event.metrics %}
**相關指標**:
{% for key, value in event.metrics.items() %}
- {{ key }}: {{ value }}
{% endfor %}
{% endif %}

---

{% endfor %}

## {{ t('root_cause_analysis') }}

{% if root_cause %}
### 主要原因

{{ root_cause.primary | default('主要原因待分析') }}

### 輔助原因

{% for cause in root_cause.contributing | default([]) %}
- {{ cause }}
{% endfor %}

### 根本原因

{{ root_cause.root | default('根本原因待深入分析') }}

### 技術細節

```
{{ root_cause.technical_details | default('技術細節待補充') }}
```

{% else %}
根因分析待補充
{% endif %}

## {{ t('action_items') }}

{% if action_items %}
| {{ t('priority') }} | {{ t('description') }} | {{ t('assignee') }} | {{ t('due_date') }} | {{ t('status') }} |
|----------|-------------|----------|------------|---------|
{% for item in action_items %}
| {{ item.priority | default('Medium') }} | {{ item.description }} | {{ item.assignee | default('未指派') }} | {{ item.due_date | format_timestamp(language) if item.due_date else '未設定' }} | {{ item.status | default('待處理') }} |
{% endfor %}

### 分類統計

{% set high_priority = action_items | selectattr('priority', 'equalto', 'High') | list %}
{% set medium_priority = action_items | selectattr('priority', 'equalto', 'Medium') | list %}
{% set low_priority = action_items | selectattr('priority', 'equalto', 'Low') | list %}

- **高優先級**: {{ high_priority | length }} 項
- **中優先級**: {{ medium_priority | length }} 項  
- **低優先級**: {{ low_priority | length }} 項

{% else %}
暫無行動項目
{% endif %}

## {{ t('lessons_learned') }}

{% if lessons_learned %}
### 正面經驗

{% for lesson in lessons_learned.positive | default([]) %}
- {{ lesson }}
{% endfor %}

### 改進機會

{% for improvement in lessons_learned.improvements | default([]) %}
- {{ improvement }}
{% endfor %}

### 流程優化

{% for process in lessons_learned.process | default([]) %}
- {{ process }}
{% endfor %}

### 技術優化

{% for tech in lessons_learned.technical | default([]) %}
- {{ tech }}
{% endfor %}

{% else %}
經驗教訓待總結
{% endif %}

## {{ t('appendix') }}

{% if appendix %}

### 相關文檔

{% for doc in appendix.documents | default([]) %}
- [{{ doc.title }}]({{ doc.url }})
{% endfor %}

### 參考資料

{% for ref in appendix.references | default([]) %}
- {{ ref }}
{% endfor %}

### 技術配置

{% if appendix.config %}
```yaml
{{ appendix.config }}
```
{% endif %}

{% endif %}

---

### 表格數據

{% for table in table_embeddings | default([]) %}
{{ table }}
{% endfor %}

---

**報告完成時間**: {{ now | format_timestamp(language) }}  
**報告負責人**: {{ report_owner | default('Detectviz 系統') }}  
**下次複查時間**: {{ next_review_date | format_timestamp(language) if next_review_date else '待安排' }}

{% if report_meta.signature %}
**電子簽名**: {{ report_meta.signature }}
{% endif %}