import { db } from '@/lib/db';
import { NextResponse } from 'next/server';

export async function GET() {
  try {
    // 获取所有云账户
    const accounts = await db.cloudAccount.findMany();
    
    // 获取本月和上月的时间范围
    const now = new Date();
    const thisMonthStart = new Date(now.getFullYear(), now.getMonth(), 1);
    const lastMonthStart = new Date(now.getFullYear(), now.getMonth() - 1, 1);
    const lastMonthEnd = new Date(now.getFullYear(), now.getMonth(), 0);
    
    // 本月成本
    const thisMonthCosts = await db.costRecord.aggregate({
      where: {
        date: { gte: thisMonthStart, lte: now },
      },
      _sum: { cost: true },
    });
    
    // 上月成本
    const lastMonthCosts = await db.costRecord.aggregate({
      where: {
        date: { gte: lastMonthStart, lte: lastMonthEnd },
      },
      _sum: { cost: true },
    });
    
    // 按分类统计本月成本
    const costsByCategory = await db.costRecord.groupBy({
      by: ['category'],
      where: {
        date: { gte: thisMonthStart, lte: now },
      },
      _sum: { cost: true },
    });
    
    // 按云账户统计本月成本
    const costsByAccount = await db.costRecord.groupBy({
      by: ['cloudAccountId'],
      where: {
        date: { gte: thisMonthStart, lte: now },
      },
      _sum: { cost: true },
    });
    
    // 资源总数
    const totalResources = await db.resource.count();
    const runningResources = await db.resource.count({
      where: { status: 'running' },
    });
    
    // 浪费检测统计
    const wasteStats = await db.wasteDetection.groupBy({
      by: ['severity'],
      where: { status: 'open' },
      _count: { id: true },
      _sum: { estimatedSavings: true },
    });
    
    // 闲置资源统计
    const idleStats = await db.idleResource.aggregate({
      where: { status: { in: ['active', 'reviewing'] } },
      _count: { id: true },
      _sum: { potentialSavings: true },
    });
    
    // 预算告警
    const budgetAlerts = await db.budgetAlert.findMany({
      where: { acknowledged: false },
      orderBy: { triggeredAt: 'desc' },
      take: 5,
    });
    
    // 最近7天的成本趋势
    const sevenDaysAgo = new Date(now);
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
    
    const recentCosts = await db.costRecord.groupBy({
      by: ['date'],
      where: {
        date: { gte: sevenDaysAgo, lte: now },
      },
      _sum: { cost: true },
      orderBy: { date: 'asc' },
    });
    
    // 成本预测 - 使用select代替include来避免Turbopack缓存问题
    const predictionsRaw = await db.costPrediction.findMany();
    
    // 手动关联账户
    const predictions = predictionsRaw.map(p => {
      const account = accounts.find(a => a.id === p.cloudAccountId);
      return {
        account: account?.name || 'Unknown',
        month: p.predictionMonth,
        predicted: p.predictedCost,
        confidence: p.confidence,
        trend: p.trend,
      };
    });
    
    // 计算环比变化
    const thisMonthTotal = thisMonthCosts._sum.cost || 0;
    const lastMonthTotal = lastMonthCosts._sum.cost || 0;
    const changePercent = lastMonthTotal > 0 
      ? ((thisMonthTotal - lastMonthTotal) / lastMonthTotal * 100).toFixed(1)
      : 0;
    
    // 汇总浪费潜在节省
    const totalWasteSavings = wasteStats.reduce((sum, w) => sum + (w._sum.estimatedSavings || 0), 0);
    
    return NextResponse.json({
      summary: {
        totalCost: Math.round(thisMonthTotal * 100) / 100,
        lastMonthCost: Math.round(lastMonthTotal * 100) / 100,
        changePercent,
        totalResources,
        runningResources,
        wasteIssues: wasteStats.reduce((sum, w) => sum + w._count.id, 0),
        idleResources: idleStats._count.id,
        potentialSavings: Math.round((totalWasteSavings + (idleStats._sum.potentialSavings || 0)) * 100) / 100,
      },
      costsByCategory: costsByCategory.map(c => ({
        category: c.category,
        cost: Math.round((c._sum.cost || 0) * 100) / 100,
      })),
      costsByAccount: costsByAccount.map(c => {
        const account = accounts.find(a => a.id === c.cloudAccountId);
        return {
          accountId: c.cloudAccountId,
          accountName: account?.name || 'Unknown',
          cost: Math.round((c._sum.cost || 0) * 100) / 100,
          budget: account?.monthlyBudget || 0,
        };
      }),
      wasteStats: wasteStats.map(w => ({
        severity: w.severity,
        count: w._count.id,
        savings: Math.round((w._sum.estimatedSavings || 0) * 100) / 100,
      })),
      idleStats: {
        count: idleStats._count.id,
        savings: Math.round((idleStats._sum.potentialSavings || 0) * 100) / 100,
      },
      budgetAlerts,
      recentCosts: recentCosts.map(r => ({
        date: r.date.toISOString().split('T')[0],
        cost: Math.round((r._sum.cost || 0) * 100) / 100,
      })),
      predictions,
      accounts: accounts.map(a => ({
        id: a.id,
        name: a.name,
        provider: a.provider,
        status: a.status,
        budget: a.monthlyBudget,
      })),
    });
  } catch (error) {
    console.error('获取仪表盘数据失败:', error);
    return NextResponse.json({ error: '获取数据失败' }, { status: 500 });
  }
}
