import { db } from '@/lib/db';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const accountId = searchParams.get('accountId');
    
    // 构建查询条件
    const where: Record<string, unknown> = {};
    if (accountId) {
      where.cloudAccountId = accountId;
    }
    
    // 获取所有云账户
    const accounts = await db.cloudAccount.findMany();
    const accountMap = new Map(accounts.map(a => [a.id, a]));
    
    // 获取预测数据 - 不使用include来避免Turbopack缓存问题
    const predictionsRaw = await db.costPrediction.findMany({
      where,
      orderBy: { predictionMonth: 'asc' },
    });
    
    // 获取历史数据用于对比
    const now = new Date();
    const threeMonthsAgo = new Date(now);
    threeMonthsAgo.setMonth(threeMonthsAgo.getMonth() - 3);
    
    // 获取按月的成本数据
    const monthlyCosts = await db.costRecord.groupBy({
      by: ['cloudAccountId'],
      where: {
        date: { gte: threeMonthsAgo, lte: now },
      },
      _sum: { cost: true },
    });
    
    // 计算平均月度成本
    const avgMonthlyCosts = new Map<string, number>();
    for (const mc of monthlyCosts) {
      avgMonthlyCosts.set(mc.cloudAccountId, (mc._sum.cost || 0) / 3);
    }
    
    // 获取过去30天的每日成本用于趋势分析
    const thirtyDaysAgo = new Date(now);
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
    
    const dailyCosts = await db.costRecord.groupBy({
      by: ['date', 'cloudAccountId'],
      where: {
        date: { gte: thirtyDaysAgo, lte: now },
      },
      _sum: { cost: true },
      orderBy: { date: 'asc' },
    });
    
    // 按账户整理每日成本
    const dailyCostsByAccount = new Map<string, Array<{ date: string; cost: number }>>();
    for (const dc of dailyCosts) {
      const existing = dailyCostsByAccount.get(dc.cloudAccountId) || [];
      existing.push({
        date: dc.date.toISOString().split('T')[0],
        cost: dc._sum.cost || 0,
      });
      dailyCostsByAccount.set(dc.cloudAccountId, existing);
    }
    
    // 格式化预测数据
    const formattedPredictions = predictionsRaw.map(p => {
      const avgCost = avgMonthlyCosts.get(p.cloudAccountId) || 0;
      const account = accountMap.get(p.cloudAccountId);
      const budget = account?.monthlyBudget || 0;
      const budgetUtilization = budget > 0 ? (p.predictedCost / budget * 100) : 0;
      
      return {
        id: p.id,
        account: {
          id: p.cloudAccountId,
          name: account?.name || 'Unknown',
          provider: account?.provider || 'unknown',
          budget: account?.monthlyBudget || 0,
        },
        month: p.predictionMonth,
        predictedCost: p.predictedCost,
        confidence: p.confidence,
        trend: p.trend,
        factors: p.factors ? JSON.parse(p.factors) : null,
        avgHistoricalCost: Math.round(avgCost * 100) / 100,
        budgetUtilization: Math.round(budgetUtilization * 10) / 10,
        overBudget: p.predictedCost > budget,
        budgetGap: Math.round((p.predictedCost - budget) * 100) / 100,
      };
    });
    
    // 总体预测汇总
    const totalPredicted = predictionsRaw.reduce((sum, p) => sum + p.predictedCost, 0);
    const totalBudget = predictionsRaw.reduce((sum, p) => {
      const account = accountMap.get(p.cloudAccountId);
      return sum + (account?.monthlyBudget || 0);
    }, 0);
    
    // 趋势分析
    const trendCounts = {
      increasing: predictionsRaw.filter(p => p.trend === 'increasing').length,
      decreasing: predictionsRaw.filter(p => p.trend === 'decreasing').length,
      stable: predictionsRaw.filter(p => p.trend === 'stable').length,
    };
    
    // 按云服务商统计预测
    const byProvider = predictionsRaw.reduce((acc, p) => {
      const account = accountMap.get(p.cloudAccountId);
      const provider = account?.provider || 'unknown';
      if (!acc[provider]) {
        acc[provider] = { count: 0, predicted: 0, budget: 0 };
      }
      acc[provider].count++;
      acc[provider].predicted += p.predictedCost;
      acc[provider].budget += account?.monthlyBudget || 0;
      return acc;
    }, {} as Record<string, { count: number; predicted: number; budget: number }>);
    
    return NextResponse.json({
      predictions: formattedPredictions,
      summary: {
        totalPredicted: Math.round(totalPredicted * 100) / 100,
        totalBudget: Math.round(totalBudget * 100) / 100,
        overBudgetCount: formattedPredictions.filter(p => p.overBudget).length,
        avgConfidence: predictionsRaw.length > 0
          ? Math.round(predictionsRaw.reduce((sum, p) => sum + p.confidence, 0) / predictionsRaw.length * 100) / 100
          : 0,
        trends: trendCounts,
      },
      byProvider: Object.entries(byProvider).map(([provider, data]) => ({
        provider,
        count: data.count,
        predicted: Math.round(data.predicted * 100) / 100,
        budget: Math.round(data.budget * 100) / 100,
        utilization: data.budget > 0 ? Math.round(data.predicted / data.budget * 100) : 0,
      })),
      dailyTrends: Object.fromEntries(dailyCostsByAccount),
    });
  } catch (error) {
    console.error('获取预测数据失败:', error);
    return NextResponse.json({ error: '获取数据失败' }, { status: 500 });
  }
}
