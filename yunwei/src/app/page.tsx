'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  Legend,
  ComposedChart,
  Line,
} from 'recharts';
import {
  DollarSign,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  Zap,
  Server,
  Database,
  Network,
  HardDrive,
  Clock,
  ArrowUpRight,
  ArrowDownRight,
  RefreshCw,
  Download,
  Settings,
  Bell,
  CheckCircle2,
  XCircle,
  Pause,
  Play,
  MoreHorizontal,
  Target,
  PiggyBank,
  Activity,
  Cloud,
  Container,
} from 'lucide-react';
import { toast } from 'sonner';

// 类型定义
interface DashboardData {
  summary: {
    totalCost: number;
    lastMonthCost: number;
    changePercent: string;
    totalResources: number;
    runningResources: number;
    wasteIssues: number;
    idleResources: number;
    potentialSavings: number;
  };
  costsByCategory: Array<{ category: string; cost: number }>;
  costsByAccount: Array<{
    accountId: string;
    accountName: string;
    cost: number;
    budget: number;
  }>;
  wasteStats: Array<{ severity: string; count: number; savings: number }>;
  idleStats: { count: number; savings: number };
  budgetAlerts: Array<{
    id: string;
    alertType: string;
    message: string;
    triggeredAt: string;
    threshold: number;
  }>;
  recentCosts: Array<{ date: string; cost: number }>;
  predictions: Array<{
    account: string;
    month: string;
    predicted: number;
    confidence: number;
    trend: string;
  }>;
  accounts: Array<{
    id: string;
    name: string;
    provider: string;
    status: string;
    budget: number | null;
  }>;
}

interface CostData {
  summary: {
    totalCost: number;
    avgDailyCost: number;
    trendPercent: string;
    dateRange: string;
  };
  timeline: Array<Record<string, unknown>>;
  byService: Array<{ service: string; category: string; cost: number }>;
  byCategory: Array<{ category: string; cost: number }>;
}

interface WasteData {
  detections: Array<{
    id: string;
    resource: {
      name: string;
      type: string;
      account: string;
      provider: string;
    };
    wasteType: string;
    severity: string;
    estimatedSavings: number;
    reason: string;
    recommendation: string;
    status: string;
    detectedAt: string;
  }>;
  summary: {
    total: number;
    totalSavings: number;
    bySeverity: Array<{ severity: string; count: number; savings: number }>;
    byType: Array<{ wasteType: string; count: number; savings: number }>;
  };
}

interface IdleData {
  resources: Array<{
    id: string;
    resource: {
      name: string;
      type: string;
      category: string;
      account: string;
      provider: string;
    };
    idleType: string;
    avgCpuUsage: number;
    avgMemoryUsage: number;
    idleDays: number;
    monthlyCost: number;
    potentialSavings: number;
    recommendation: string;
    status: string;
    detectedAt: string;
  }>;
  summary: {
    total: number;
    totalSavings: number;
    totalMonthlyCost: number;
    byType: Array<{ idleType: string; count: number; savings: number }>;
    idleDaysDistribution: Array<{ range: string; count: number; savings: number }>;
  };
}

interface PredictionData {
  predictions: Array<{
    id: string;
    account: {
      name: string;
      provider: string;
      budget: number | null;
    };
    month: string;
    predictedCost: number;
    confidence: number;
    trend: string;
    avgHistoricalCost: number;
    budgetUtilization: number;
    overBudget: boolean;
    budgetGap: number;
  }>;
  summary: {
    totalPredicted: number;
    totalBudget: number;
    overBudgetCount: number;
    avgConfidence: number;
    trends: { increasing: number; decreasing: number; stable: number };
  };
  byProvider: Array<{
    provider: string;
    count: number;
    predicted: number;
    budget: number;
    utilization: number;
  }>;
}

// 颜色配置
const COLORS = {
  compute: '#ef4444',
  storage: '#f97316',
  network: '#22c55e',
  database: '#06b6d4',
  other: '#8b5cf6',
};

const SEVERITY_COLORS = {
  critical: '#dc2626',
  high: '#f97316',
  medium: '#eab308',
  low: '#22c55e',
};

const PROVIDER_ICONS: Record<string, React.ReactNode> = {
  aws: <Cloud className="h-4 w-4" />,
  gcp: <Cloud className="h-4 w-4" />,
  kubernetes: <Container className="h-4 w-4" />,
};

export default function CostControlDashboard() {
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null);
  const [costData, setCostData] = useState<CostData | null>(null);
  const [wasteData, setWasteData] = useState<WasteData | null>(null);
  const [idleData, setIdleData] = useState<IdleData | null>(null);
  const [predictionData, setPredictionData] = useState<PredictionData | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');
  const [selectedAccount, setSelectedAccount] = useState<string>('all');
  const [dateRange, setDateRange] = useState<string>('30d');

  // 初始化数据
  const initData = useCallback(async () => {
    try {
      const res = await fetch('/api/init');
      const data = await res.json();
      console.log('Init result:', data);
    } catch (error) {
      console.error('Init error:', error);
    }
  }, []);

  // 加载仪表盘数据
  const loadDashboard = useCallback(async () => {
    try {
      const res = await fetch('/api/dashboard');
      const data = await res.json();
      setDashboardData(data);
    } catch (error) {
      console.error('Load dashboard error:', error);
      toast.error('加载数据失败');
    }
  }, []);

  // 加载成本数据
  const loadCosts = useCallback(async () => {
    try {
      const params = new URLSearchParams();
      if (selectedAccount !== 'all') {
        params.set('accountId', selectedAccount);
      }
      params.set('range', dateRange);
      
      const res = await fetch(`/api/costs?${params}`);
      const data = await res.json();
      setCostData(data);
    } catch (error) {
      console.error('Load costs error:', error);
    }
  }, [selectedAccount, dateRange]);

  // 加载浪费检测数据
  const loadWaste = useCallback(async () => {
    try {
      const res = await fetch('/api/waste');
      const data = await res.json();
      setWasteData(data);
    } catch (error) {
      console.error('Load waste error:', error);
    }
  }, []);

  // 加载闲置资源数据
  const loadIdle = useCallback(async () => {
    try {
      const res = await fetch('/api/idle');
      const data = await res.json();
      setIdleData(data);
    } catch (error) {
      console.error('Load idle error:', error);
    }
  }, []);

  // 加载预测数据
  const loadPredictions = useCallback(async () => {
    try {
      const res = await fetch('/api/predictions');
      const data = await res.json();
      setPredictionData(data);
    } catch (error) {
      console.error('Load predictions error:', error);
    }
  }, []);

  // 更新浪费检测状态
  const updateWasteStatus = async (id: string, status: string) => {
    try {
      await fetch('/api/waste', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, status }),
      });
      toast.success('状态已更新');
      loadWaste();
    } catch {
      toast.error('更新失败');
    }
  };

  // 更新闲置资源状态
  const updateIdleStatus = async (id: string, status: string) => {
    try {
      await fetch('/api/idle', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, status }),
      });
      toast.success('状态已更新');
      loadIdle();
    } catch {
      toast.error('更新失败');
    }
  };

  // 初始加载
  useEffect(() => {
    const init = async () => {
      setLoading(true);
      await initData();
      await loadDashboard();
      setLoading(false);
    };
    init();
  }, []);

  // Tab切换时加载数据
  useEffect(() => {
    const loadTabData = async () => {
      if (activeTab === 'costs' && !costData) {
        await loadCosts();
      } else if (activeTab === 'waste' && !wasteData) {
        await loadWaste();
      } else if (activeTab === 'idle' && !idleData) {
        await loadIdle();
      } else if (activeTab === 'predictions' && !predictionData) {
        await loadPredictions();
      }
    };
    loadTabData();
  }, [activeTab]);

  // 格式化金额
  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('zh-CN', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  // 格式化百分比
  const formatPercent = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(1)}%`;
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-950 dark:to-slate-900">
        <div className="text-center">
          <RefreshCw className="h-12 w-12 animate-spin text-primary mx-auto mb-4" />
          <p className="text-muted-foreground">正在加载数据...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-950 dark:to-slate-900">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b bg-background/80 backdrop-blur-sm">
        <div className="container mx-auto px-4 py-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center">
                <PiggyBank className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold">云成本控制中心</h1>
                <p className="text-xs text-muted-foreground">Cloud Cost Control System</p>
              </div>
            </div>
            
            <div className="flex items-center gap-3">
              <Select value={selectedAccount} onValueChange={setSelectedAccount}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="选择账户" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部账户</SelectItem>
                  {dashboardData?.accounts.map((account) => (
                    <SelectItem key={account.id} value={account.id}>
                      {account.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              
              <Button variant="outline" size="icon" onClick={() => loadDashboard()}>
                <RefreshCw className="h-4 w-4" />
              </Button>
              
              <Button variant="outline" size="icon">
                <Bell className="h-4 w-4" />
              </Button>
              
              <Button variant="outline" size="icon">
                <Settings className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-6">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="grid w-full grid-cols-5 lg:w-auto lg:inline-grid">
            <TabsTrigger value="overview" className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              <span className="hidden sm:inline">总览</span>
            </TabsTrigger>
            <TabsTrigger value="costs" className="flex items-center gap-2">
              <DollarSign className="h-4 w-4" />
              <span className="hidden sm:inline">成本统计</span>
            </TabsTrigger>
            <TabsTrigger value="predictions" className="flex items-center gap-2">
              <Target className="h-4 w-4" />
              <span className="hidden sm:inline">成本预测</span>
            </TabsTrigger>
            <TabsTrigger value="waste" className="flex items-center gap-2">
              <AlertTriangle className="h-4 w-4" />
              <span className="hidden sm:inline">浪费检测</span>
            </TabsTrigger>
            <TabsTrigger value="idle" className="flex items-center gap-2">
              <Pause className="h-4 w-4" />
              <span className="hidden sm:inline">闲置资源</span>
            </TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="space-y-6">
            {/* Summary Cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
              <Card className="relative overflow-hidden">
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">本月成本</CardTitle>
                  <DollarSign className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{formatCurrency(dashboardData?.summary.totalCost || 0)}</div>
                  <div className="flex items-center gap-1 mt-1">
                    {parseFloat(dashboardData?.summary.changePercent || '0') >= 0 ? (
                      <ArrowUpRight className="h-4 w-4 text-red-500" />
                    ) : (
                      <ArrowDownRight className="h-4 w-4 text-green-500" />
                    )}
                    <span className={parseFloat(dashboardData?.summary.changePercent || '0') >= 0 ? 'text-red-500 text-sm' : 'text-green-500 text-sm'}>
                      {formatPercent(parseFloat(dashboardData?.summary.changePercent || '0'))} vs 上月
                    </span>
                  </div>
                </CardContent>
                <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-red-500 to-orange-500" />
              </Card>

              <Card className="relative overflow-hidden">
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">活跃资源</CardTitle>
                  <Server className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {dashboardData?.summary.runningResources || 0}
                    <span className="text-muted-foreground text-sm font-normal"> / {dashboardData?.summary.totalResources || 0}</span>
                  </div>
                  <Progress 
                    value={dashboardData ? (dashboardData.summary.runningResources / dashboardData.summary.totalResources * 100) : 0} 
                    className="mt-2"
                  />
                </CardContent>
                <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-blue-500 to-cyan-500" />
              </Card>

              <Card className="relative overflow-hidden">
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">浪费问题</CardTitle>
                  <AlertTriangle className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-amber-600">{dashboardData?.summary.wasteIssues || 0}</div>
                  <p className="text-xs text-muted-foreground mt-1">待处理优化建议</p>
                </CardContent>
                <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-amber-500 to-yellow-500" />
              </Card>

              <Card className="relative overflow-hidden">
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">潜在节省</CardTitle>
                  <PiggyBank className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-green-600">{formatCurrency(dashboardData?.summary.potentialSavings || 0)}</div>
                  <p className="text-xs text-muted-foreground mt-1">月度预估节省</p>
                </CardContent>
                <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-green-500 to-emerald-500" />
              </Card>
            </div>

            {/* Charts Row */}
            <div className="grid gap-4 lg:grid-cols-2">
              {/* Cost Trend Chart */}
              <Card>
                <CardHeader>
                  <CardTitle>成本趋势</CardTitle>
                  <CardDescription>近7天成本变化</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="h-[300px]">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={dashboardData?.recentCosts || []}>
                        <defs>
                          <linearGradient id="colorCost" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3} />
                            <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                        <XAxis 
                          dataKey="date" 
                          tickFormatter={(v) => v.slice(5)}
                          className="text-xs"
                        />
                        <YAxis 
                          tickFormatter={(v) => `$${v}`}
                          className="text-xs"
                        />
                        <Tooltip 
                          formatter={(value: number) => formatCurrency(value)}
                          labelFormatter={(label) => `日期: ${label}`}
                        />
                        <Area
                          type="monotone"
                          dataKey="cost"
                          stroke="#ef4444"
                          strokeWidth={2}
                          fillOpacity={1}
                          fill="url(#colorCost)"
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </CardContent>
              </Card>

              {/* Cost by Category */}
              <Card>
                <CardHeader>
                  <CardTitle>成本分布</CardTitle>
                  <CardDescription>按资源类型分类</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="h-[300px]">
                    <ResponsiveContainer width="100%" height="100%">
                      <PieChart>
                        <Pie
                          data={dashboardData?.costsByCategory || []}
                          cx="50%"
                          cy="50%"
                          innerRadius={60}
                          outerRadius={100}
                          paddingAngle={5}
                          dataKey="cost"
                          nameKey="category"
                        >
                          {dashboardData?.costsByCategory.map((entry, index) => (
                            <Cell 
                              key={`cell-${index}`} 
                              fill={COLORS[entry.category as keyof typeof COLORS] || '#8b5cf6'} 
                            />
                          ))}
                        </Pie>
                        <Tooltip formatter={(value: number) => formatCurrency(value)} />
                        <Legend />
                      </PieChart>
                    </ResponsiveContainer>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Budget and Predictions */}
            <div className="grid gap-4 lg:grid-cols-2">
              {/* Account Budget */}
              <Card>
                <CardHeader>
                  <CardTitle>账户预算使用</CardTitle>
                  <CardDescription>各云账户预算消耗情况</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {dashboardData?.costsByAccount.map((account) => {
                    const utilization = (account.cost / account.budget) * 100;
                    return (
                      <div key={account.accountId} className="space-y-2">
                        <div className="flex justify-between text-sm">
                          <span className="font-medium">{account.accountName}</span>
                          <span className="text-muted-foreground">
                            {formatCurrency(account.cost)} / {formatCurrency(account.budget)}
                          </span>
                        </div>
                        <Progress 
                          value={utilization} 
                          className={`h-2 ${utilization > 90 ? 'bg-red-100' : utilization > 70 ? 'bg-amber-100' : ''}`}
                        />
                        {utilization > 90 && (
                          <p className="text-xs text-red-600">预算即将用尽，请关注！</p>
                        )}
                      </div>
                    );
                  })}
                </CardContent>
              </Card>

              {/* Predictions Summary */}
              <Card>
                <CardHeader>
                  <CardTitle>成本预测</CardTitle>
                  <CardDescription>下月成本预测</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {dashboardData?.predictions.map((pred, index) => (
                    <div key={index} className="flex items-center justify-between p-3 rounded-lg bg-muted/50">
                      <div>
                        <p className="font-medium">{pred.account}</p>
                        <p className="text-sm text-muted-foreground">{pred.month}</p>
                      </div>
                      <div className="text-right">
                        <p className="font-bold">{formatCurrency(pred.predicted)}</p>
                        <Badge variant={pred.trend === 'increasing' ? 'destructive' : pred.trend === 'decreasing' ? 'default' : 'secondary'}>
                          {pred.trend === 'increasing' ? <TrendingUp className="h-3 w-3 mr-1" /> : 
                           pred.trend === 'decreasing' ? <TrendingDown className="h-3 w-3 mr-1" /> : null}
                          {pred.trend === 'increasing' ? '上升' : pred.trend === 'decreasing' ? '下降' : '稳定'}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            </div>

            {/* Budget Alerts */}
            {dashboardData?.budgetAlerts && dashboardData.budgetAlerts.length > 0 && (
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Bell className="h-5 w-5" />
                    预算告警
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {dashboardData.budgetAlerts.map((alert) => (
                      <div 
                        key={alert.id}
                        className="flex items-center gap-3 p-3 rounded-lg border border-amber-200 bg-amber-50 dark:border-amber-900 dark:bg-amber-950"
                      >
                        <AlertTriangle className="h-5 w-5 text-amber-600" />
                        <div className="flex-1">
                          <p className="font-medium">{alert.message}</p>
                          <p className="text-sm text-muted-foreground">
                            {new Date(alert.triggeredAt).toLocaleString('zh-CN')}
                          </p>
                        </div>
                        <Button size="sm" variant="outline">确认</Button>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* Costs Tab */}
          <TabsContent value="costs" className="space-y-6">
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-4">
                <Select value={dateRange} onValueChange={setDateRange}>
                  <SelectTrigger className="w-[120px]">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="7d">近7天</SelectItem>
                    <SelectItem value="30d">近30天</SelectItem>
                    <SelectItem value="90d">近90天</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Button variant="outline" onClick={() => loadCosts()}>
                <RefreshCw className="h-4 w-4 mr-2" />
                刷新
              </Button>
            </div>

            {costData && (
              <>
                {/* Cost Summary */}
                <div className="grid gap-4 md:grid-cols-3">
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">总成本</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{formatCurrency(costData.summary.totalCost)}</div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">日均成本</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{formatCurrency(costData.summary.avgDailyCost)}</div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">趋势变化</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className={`text-2xl font-bold ${parseFloat(costData.summary.trendPercent) >= 0 ? 'text-red-600' : 'text-green-600'}`}>
                        {formatPercent(parseFloat(costData.summary.trendPercent))}
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Timeline Chart */}
                <Card>
                  <CardHeader>
                    <CardTitle>成本趋势分析</CardTitle>
                    <CardDescription>按分类展示成本变化</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="h-[400px]">
                      <ResponsiveContainer width="100%" height="100%">
                        <ComposedChart data={costData.timeline}>
                          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                          <XAxis dataKey="date" tickFormatter={(v) => v.slice(5)} className="text-xs" />
                          <YAxis tickFormatter={(v) => `$${v}`} className="text-xs" />
                          <Tooltip formatter={(value: number) => formatCurrency(value)} />
                          <Legend />
                          <Area type="monotone" dataKey="compute" stackId="1" stroke={COLORS.compute} fill={COLORS.compute} fillOpacity={0.6} name="计算" />
                          <Area type="monotone" dataKey="storage" stackId="1" stroke={COLORS.storage} fill={COLORS.storage} fillOpacity={0.6} name="存储" />
                          <Area type="monotone" dataKey="network" stackId="1" stroke={COLORS.network} fill={COLORS.network} fillOpacity={0.6} name="网络" />
                          <Area type="monotone" dataKey="database" stackId="1" stroke={COLORS.database} fill={COLORS.database} fillOpacity={0.6} name="数据库" />
                          <Line type="monotone" dataKey="total" stroke="#1f2937" strokeWidth={2} dot={false} name="总计" />
                        </ComposedChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>

                {/* Service Costs */}
                <div className="grid gap-4 lg:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle>服务成本TOP 10</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="h-[300px]">
                        <ResponsiveContainer width="100%" height="100%">
                          <BarChart data={costData.byService.slice(0, 10)} layout="vertical">
                            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                            <XAxis type="number" tickFormatter={(v) => `$${v}`} className="text-xs" />
                            <YAxis dataKey="service" type="category" width={100} className="text-xs" />
                            <Tooltip formatter={(value: number) => formatCurrency(value)} />
                            <Bar dataKey="cost" fill="#ef4444" radius={[0, 4, 4, 0]} />
                          </BarChart>
                        </ResponsiveContainer>
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>分类汇总</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        {costData.byCategory.map((cat) => (
                          <div key={cat.category} className="space-y-2">
                            <div className="flex justify-between">
                              <span className="capitalize">{cat.category}</span>
                              <span className="font-medium">{formatCurrency(cat.cost)}</span>
                            </div>
                            <Progress 
                              value={(cat.cost / costData.summary.totalCost) * 100} 
                              className="h-2"
                            />
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                </div>
              </>
            )}
          </TabsContent>

          {/* Predictions Tab */}
          <TabsContent value="predictions" className="space-y-6">
            {predictionData && (
              <>
                {/* Summary Cards */}
                <div className="grid gap-4 md:grid-cols-4">
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">预测总额</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{formatCurrency(predictionData.summary.totalPredicted)}</div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">预算总额</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{formatCurrency(predictionData.summary.totalBudget)}</div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">超预算账户</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold text-red-600">{predictionData.summary.overBudgetCount}</div>
                    </CardContent>
                  </Card>
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">平均置信度</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{(predictionData.summary.avgConfidence * 100).toFixed(0)}%</div>
                    </CardContent>
                  </Card>
                </div>

                {/* Predictions Detail */}
                <div className="grid gap-4 lg:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle>账户预测详情</CardTitle>
                      <CardDescription>各账户下月成本预测</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-4">
                        {predictionData.predictions.map((pred) => (
                          <div 
                            key={pred.id}
                            className={`p-4 rounded-lg border ${pred.overBudget ? 'border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950' : 'bg-muted/50'}`}
                          >
                            <div className="flex justify-between items-start mb-2">
                              <div>
                                <p className="font-medium">{pred.account.name}</p>
                                <p className="text-sm text-muted-foreground">{pred.month}</p>
                              </div>
                              <Badge variant={pred.overBudget ? 'destructive' : 'secondary'}>
                                {pred.overBudget ? '超预算' : '正常'}
                              </Badge>
                            </div>
                            <div className="grid grid-cols-2 gap-4 mt-3">
                              <div>
                                <p className="text-xs text-muted-foreground">预测成本</p>
                                <p className="font-bold">{formatCurrency(pred.predictedCost)}</p>
                              </div>
                              <div>
                                <p className="text-xs text-muted-foreground">预算</p>
                                <p className="font-bold">{formatCurrency(pred.account.budget || 0)}</p>
                              </div>
                              <div>
                                <p className="text-xs text-muted-foreground">置信度</p>
                                <p className="font-medium">{(pred.confidence * 100).toFixed(0)}%</p>
                              </div>
                              <div>
                                <p className="text-xs text-muted-foreground">趋势</p>
                                <Badge variant={pred.trend === 'increasing' ? 'destructive' : pred.trend === 'decreasing' ? 'default' : 'outline'}>
                                  {pred.trend === 'increasing' ? '↑ 上升' : pred.trend === 'decreasing' ? '↓ 下降' : '→ 稳定'}
                                </Badge>
                              </div>
                            </div>
                            {pred.overBudget && (
                              <div className="mt-3 p-2 rounded bg-red-100 dark:bg-red-900 text-sm">
                                预计超预算 {formatCurrency(pred.budgetGap)}
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>云服务商分布</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="h-[300px]">
                        <ResponsiveContainer width="100%" height="100%">
                          <BarChart data={predictionData.byProvider}>
                            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                            <XAxis dataKey="provider" className="text-xs" />
                            <YAxis tickFormatter={(v) => `$${v/1000}k`} className="text-xs" />
                            <Tooltip formatter={(value: number) => formatCurrency(value)} />
                            <Legend />
                            <Bar dataKey="predicted" name="预测" fill="#ef4444" />
                            <Bar dataKey="budget" name="预算" fill="#22c55e" />
                          </BarChart>
                        </ResponsiveContainer>
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Trend Analysis */}
                <Card>
                  <CardHeader>
                    <CardTitle>趋势分析</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-3 gap-4">
                      <div className="text-center p-4 rounded-lg bg-red-50 dark:bg-red-950">
                        <TrendingUp className="h-8 w-8 mx-auto mb-2 text-red-600" />
                        <p className="text-2xl font-bold">{predictionData.summary.trends.increasing}</p>
                        <p className="text-sm text-muted-foreground">上升趋势</p>
                      </div>
                      <div className="text-center p-4 rounded-lg bg-green-50 dark:bg-green-950">
                        <TrendingDown className="h-8 w-8 mx-auto mb-2 text-green-600" />
                        <p className="text-2xl font-bold">{predictionData.summary.trends.decreasing}</p>
                        <p className="text-sm text-muted-foreground">下降趋势</p>
                      </div>
                      <div className="text-center p-4 rounded-lg bg-muted/50">
                        <Activity className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                        <p className="text-2xl font-bold">{predictionData.summary.trends.stable}</p>
                        <p className="text-sm text-muted-foreground">稳定</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </>
            )}
          </TabsContent>

          {/* Waste Detection Tab */}
          <TabsContent value="waste" className="space-y-6">
            {wasteData && (
              <>
                {/* Summary */}
                <div className="grid gap-4 md:grid-cols-4">
                  <Card className="border-l-4 border-l-red-500">
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">问题总数</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{wasteData.summary.total}</div>
                    </CardContent>
                  </Card>
                  <Card className="border-l-4 border-l-green-500">
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">潜在节省</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold text-green-600">{formatCurrency(wasteData.summary.totalSavings)}</div>
                    </CardContent>
                  </Card>
                  {wasteData.summary.bySeverity.map((s) => (
                    <Card key={s.severity} className="border-l-4" style={{ borderLeftColor: SEVERITY_COLORS[s.severity as keyof typeof SEVERITY_COLORS] }}>
                      <CardHeader className="pb-2">
                        <CardTitle className="text-sm text-muted-foreground capitalize">{s.severity}</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="text-2xl font-bold">{s.count}</div>
                        <p className="text-sm text-muted-foreground">{formatCurrency(s.savings)} 可节省</p>
                      </CardContent>
                    </Card>
                  ))}
                </div>

                {/* By Type Chart */}
                <Card>
                  <CardHeader>
                    <CardTitle>浪费类型分布</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="h-[200px]">
                      <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={wasteData.summary.byType}>
                          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                          <XAxis dataKey="wasteType" className="text-xs" />
                          <YAxis yAxisId="left" className="text-xs" />
                          <YAxis yAxisId="right" orientation="right" className="text-xs" />
                          <Tooltip />
                          <Bar yAxisId="left" dataKey="count" name="数量" fill="#ef4444" />
                          <Bar yAxisId="right" dataKey="savings" name="节省金额" fill="#22c55e" />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>

                {/* Detections List */}
                <Card>
                  <CardHeader>
                    <CardTitle>检测详情</CardTitle>
                    <CardDescription>点击处理各项浪费问题</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <ScrollArea className="h-[500px]">
                      <div className="space-y-3 pr-4">
                        {wasteData.detections.map((detection) => (
                          <div 
                            key={detection.id}
                            className="p-4 rounded-lg border bg-card hover:shadow-md transition-shadow"
                          >
                            <div className="flex items-start justify-between">
                              <div className="flex items-start gap-3">
                                <div 
                                  className="p-2 rounded-lg"
                                  style={{ backgroundColor: `${SEVERITY_COLORS[detection.severity as keyof typeof SEVERITY_COLORS]}20` }}
                                >
                                  <AlertTriangle 
                                    className="h-5 w-5"
                                    style={{ color: SEVERITY_COLORS[detection.severity as keyof typeof SEVERITY_COLORS] }}
                                  />
                                </div>
                                <div>
                                  <div className="flex items-center gap-2">
                                    <p className="font-medium">{detection.resource.name}</p>
                                    <Badge 
                                      variant="outline"
                                      style={{ 
                                        borderColor: SEVERITY_COLORS[detection.severity as keyof typeof SEVERITY_COLORS],
                                        color: SEVERITY_COLORS[detection.severity as keyof typeof SEVERITY_COLORS]
                                      }}
                                    >
                                      {detection.severity}
                                    </Badge>
                                  </div>
                                  <p className="text-sm text-muted-foreground">
                                    {PROVIDER_ICONS[detection.resource.provider]} {detection.resource.account} · {detection.resource.type}
                                  </p>
                                  <p className="text-sm mt-2">{detection.reason}</p>
                                  <p className="text-sm text-blue-600 mt-1">💡 {detection.recommendation}</p>
                                </div>
                              </div>
                              <div className="text-right">
                                <p className="font-bold text-green-600">{formatCurrency(detection.estimatedSavings)}</p>
                                <p className="text-xs text-muted-foreground">月度节省</p>
                                <AlertDialog>
                                  <AlertDialogTrigger asChild>
                                    <Button size="sm" className="mt-2">处理</Button>
                                  </AlertDialogTrigger>
                                  <AlertDialogContent>
                                    <AlertDialogHeader>
                                      <AlertDialogTitle>确认处理</AlertDialogTitle>
                                      <AlertDialogDescription>
                                        确定要处理此浪费问题吗？此操作将标记问题为已解决。
                                      </AlertDialogDescription>
                                    </AlertDialogHeader>
                                    <AlertDialogFooter>
                                      <AlertDialogCancel>取消</AlertDialogCancel>
                                      <AlertDialogAction onClick={() => updateWasteStatus(detection.id, 'resolved')}>
                                        确认解决
                                      </AlertDialogAction>
                                    </AlertDialogFooter>
                                  </AlertDialogContent>
                                </AlertDialog>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </ScrollArea>
                  </CardContent>
                </Card>
              </>
            )}
          </TabsContent>

          {/* Idle Resources Tab */}
          <TabsContent value="idle" className="space-y-6">
            {idleData && (
              <>
                {/* Summary */}
                <div className="grid gap-4 md:grid-cols-4">
                  <Card className="border-l-4 border-l-amber-500">
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">闲置资源</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">{idleData.summary.total}</div>
                    </CardContent>
                  </Card>
                  <Card className="border-l-4 border-l-green-500">
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">潜在节省</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold text-green-600">{formatCurrency(idleData.summary.totalSavings)}</div>
                    </CardContent>
                  </Card>
                  <Card className="border-l-4 border-l-red-500">
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">当前成本</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold text-red-600">{formatCurrency(idleData.summary.totalMonthlyCost)}</div>
                    </CardContent>
                  </Card>
                  <Card className="border-l-4 border-l-blue-500">
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm text-muted-foreground">平均节省率</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {((idleData.summary.totalSavings / idleData.summary.totalMonthlyCost) * 100).toFixed(0)}%
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Idle Days Distribution */}
                <div className="grid gap-4 lg:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle>闲置时长分布</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="h-[250px]">
                        <ResponsiveContainer width="100%" height="100%">
                          <BarChart data={idleData.summary.idleDaysDistribution}>
                            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                            <XAxis dataKey="range" className="text-xs" />
                            <YAxis yAxisId="left" className="text-xs" />
                            <YAxis yAxisId="right" orientation="right" className="text-xs" />
                            <Tooltip />
                            <Bar yAxisId="left" dataKey="count" name="数量" fill="#f97316" />
                            <Bar yAxisId="right" dataKey="savings" name="节省" fill="#22c55e" />
                          </BarChart>
                        </ResponsiveContainer>
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader>
                      <CardTitle>闲置类型分布</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="space-y-3">
                        {idleData.summary.byType.map((t) => (
                          <div key={t.idleType} className="flex items-center justify-between p-3 rounded-lg bg-muted/50">
                            <div className="flex items-center gap-2">
                              <Badge>{t.idleType}</Badge>
                              <span className="text-sm">{t.count} 个资源</span>
                            </div>
                            <span className="font-medium text-green-600">{formatCurrency(t.savings)}</span>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Resources List */}
                <Card>
                  <CardHeader>
                    <CardTitle>闲置资源详情</CardTitle>
                    <CardDescription>识别并处理低利用率资源</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <ScrollArea className="h-[500px]">
                      <div className="space-y-3 pr-4">
                        {idleData.resources.map((resource) => (
                          <div 
                            key={resource.id}
                            className="p-4 rounded-lg border bg-card hover:shadow-md transition-shadow"
                          >
                            <div className="flex items-start justify-between">
                              <div className="flex items-start gap-3">
                                <div className="p-2 rounded-lg bg-amber-100 dark:bg-amber-950">
                                  <Pause className="h-5 w-5 text-amber-600" />
                                </div>
                                <div>
                                  <div className="flex items-center gap-2">
                                    <p className="font-medium">{resource.resource.name}</p>
                                    <Badge variant="outline">{resource.idleType}</Badge>
                                  </div>
                                  <p className="text-sm text-muted-foreground">
                                    {resource.resource.account} · {resource.resource.type}
                                  </p>
                                  <div className="grid grid-cols-3 gap-4 mt-3">
                                    <div>
                                      <p className="text-xs text-muted-foreground">CPU使用率</p>
                                      <p className="font-medium">{resource.avgCpuUsage.toFixed(1)}%</p>
                                    </div>
                                    <div>
                                      <p className="text-xs text-muted-foreground">内存使用率</p>
                                      <p className="font-medium">{resource.avgMemoryUsage.toFixed(1)}%</p>
                                    </div>
                                    <div>
                                      <p className="text-xs text-muted-foreground">闲置天数</p>
                                      <p className="font-medium">{resource.idleDays} 天</p>
                                    </div>
                                  </div>
                                  <p className="text-sm text-blue-600 mt-2">💡 建议: {resource.recommendation}</p>
                                </div>
                              </div>
                              <div className="text-right">
                                <p className="font-bold text-red-600">{formatCurrency(resource.monthlyCost)}</p>
                                <p className="text-xs text-muted-foreground">月成本</p>
                                <p className="font-medium text-green-600 mt-1">{formatCurrency(resource.potentialSavings)}</p>
                                <p className="text-xs text-muted-foreground">可节省</p>
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <Button size="sm" variant="outline" className="mt-2">
                                      操作 <MoreHorizontal className="h-4 w-4 ml-1" />
                                    </Button>
                                  </DropdownMenuTrigger>
                                  <DropdownMenuContent>
                                    <DropdownMenuItem onClick={() => updateIdleStatus(resource.id, 'reviewing')}>
                                      <Clock className="h-4 w-4 mr-2" />
                                      标记审核中
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => updateIdleStatus(resource.id, 'actioned')}>
                                      <CheckCircle2 className="h-4 w-4 mr-2" />
                                      已处理
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => updateIdleStatus(resource.id, 'dismissed')}>
                                      <XCircle className="h-4 w-4 mr-2" />
                                      忽略
                                    </DropdownMenuItem>
                                  </DropdownMenuContent>
                                </DropdownMenu>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </ScrollArea>
                  </CardContent>
                </Card>
              </>
            )}
          </TabsContent>
        </Tabs>
      </main>

      {/* Footer */}
      <footer className="border-t bg-background/80 backdrop-blur-sm mt-auto">
        <div className="container mx-auto px-4 py-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <p className="text-sm text-muted-foreground">
              © 2024 云成本控制中心 - 智能化云资源成本管理平台
            </p>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span>数据更新时间: {new Date().toLocaleString('zh-CN')}</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
