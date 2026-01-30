# Performance Monitor Command

Real-time performance monitoring dashboard with analytics, alerts, and optimization recommendations.

**Features:**
📊 **Real-Time Dashboard** - Live system performance monitoring with comprehensive metrics
📈 **Trend Analysis** - Historical performance analysis with intelligent pattern recognition
🚨 **Smart Alerts** - Configurable thresholds with severity levels and actionable recommendations  
⚡ **Performance Analytics** - CPU, memory, disk, network, and tool response time monitoring
💡 **Optimization Insights** - AI-powered recommendations for performance improvements
📋 **Historical Data** - Long-term performance tracking with baseline comparison capabilities

**Usage Examples:**
- Show dashboard: `/monitor`
- Analyze trends: `/monitor --action analyze --time-range 7d`
- Check alerts: `/monitor --action alert --cpu-threshold 70`
- Export data: `/monitor --format json --time-range 1h`
- Real-time mode: `/monitor --real-time --metrics cpu_usage,memory_usage`

**Available Metrics:**
- cpu_usage, memory_usage, disk_io, network_io
- tool_response_times, error_rates, throughput, latency
- resource_utilization, user_activity

**Alert Thresholds:**
- CPU: 80% (configurable)
- Memory: 85% (configurable)  
- Response Time: 5.0s (configurable)
- Error Rate: 5.0% (configurable)

**Time Ranges:** 1h, 6h, 24h, 7d, 30d, custom

Target: $ARGUMENTS

Monitors system and development workflow performance with real-time dashboards, intelligent alerting, historical analysis, and optimization recommendations for enhanced productivity.