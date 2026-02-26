import { db } from '@/lib/db';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const range = searchParams.get('range') || '30d'; // 7d, 30d, 90d
    const accountId = searchParams.get('accountId');
    const category = searchParams.get('category');
    
    // 计算时间范围
    const now = new Date();
    let startDate = new Date(now);
    switch (range) {
      case '7d':
        startDate.setDate(startDate.getDate() - 7);
        break;
      case '90d':
        startDate.setDate(startDate.getDate() - 90);
        break;
      default:
        startDate.setDate(startDate.getDate() - 30);
    }
    
    // 构建查询条件
    const where: Record<string, unknown> = {
      date: { gte: startDate, lte: now },
    };
    if (accountId) {
      where.cloudAccountId = accountId;
    }
    if (category) {
      where.category = category;
    }
    
    // 按日期统计成本
    const dailyCosts = await db.costRecord.groupBy({
      by: ['date', 'category'],
      where,
      _sum: { cost: true },
      orderBy: { date: 'asc' },
    });
    
    // 按服务统计成本
    const serviceCosts = await db.costRecord.groupBy({
      by: ['service', 'category'],
      where,
      _sum: { cost: true },
      orderBy: { _sum: { cost: 'desc' } },
    });
    
    // 按分类统计
    const categoryCosts = await db.costRecord.groupBy({
      by: ['category'],
      where,
      _sum: { cost: true },
      orderBy: { _sum: { cost: 'desc' } },
    });
    
    // 总成本
    const totalCost = await db.costRecord.aggregate({
      where,
      _sum: { cost: true },
    });
    
    // 格式化日期成本数据
    const dateMap = new Map<string, Record<string, number>>();
    const categories = new Set<string>();
    
    for (const dc of dailyCosts) {
      const dateStr = dc.date.toISOString().split('T')[0];
      categories.add(dc.category);
      
      if (!dateMap.has(dateStr)) {
        dateMap.set(dateStr, {});
      }
      const dateData = dateMap.get(dateStr)!;
      dateData[dc.category] = (dateData[dc.category] || 0) + (dc._sum.cost || 0);
    }
    
    // 转换为数组格式
    const timelineData = Array.from(dateMap.entries())
      .sort((a, b) => a[0].localeCompare(b[0]))
      .map(([date, costs]) => ({
        date,
        ...Object.fromEntries(
          Array.from(categories).map(cat => [
            cat,
            Math.round((costs[cat] || 0) * 100) / 100,
          ])
        ),
        total: Math.round(Object.values(costs).reduce((sum, c) => sum + c, 0) * 100) / 100,
      }));
    
    // 计算趋势
    const firstWeek = timelineData.slice(0, 7).reduce((sum, d) => sum + d.total, 0);
    const lastWeek = timelineData.slice(-7).reduce((sum, d) => sum + d.total, 0);
    const trendPercent = firstWeek > 0 
      ? ((lastWeek - firstWeek) / firstWeek * 100).toFixed(1)
      : '0';
    
    // 日均成本
    const avgDailyCost = timelineData.length > 0
      ? timelineData.reduce((sum, d) => sum + d.total, 0) / timelineData.length
      : 0;
    
    return NextResponse.json({
      summary: {
        totalCost: Math.round((totalCost._sum.cost || 0) * 100) / 100,
        avgDailyCost: Math.round(avgDailyCost * 100) / 100,
        trendPercent,
        dateRange: range,
      },
      timeline: timelineData,
      byService: serviceCosts.slice(0, 20).map(s => ({
        service: s.service,
        category: s.category,
        cost: Math.round((s._sum.cost || 0) * 100) / 100,
      })),
      byCategory: categoryCosts.map(c => ({
        category: c.category,
        cost: Math.round((c._sum.cost || 0) * 100) / 100,
      })),
    });
  } catch (error) {
    console.error('获取成本数据失败:', error);
    return NextResponse.json({ error: '获取数据失败' }, { status: 500 });
  }
}
