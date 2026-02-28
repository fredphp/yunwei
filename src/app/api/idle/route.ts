import { db } from '@/lib/db';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const idleType = searchParams.get('idleType');
    const status = searchParams.get('status') || 'active,reviewing';
    
    // 构建查询条件
    const where: Record<string, unknown> = {};
    if (idleType) {
      where.idleType = idleType;
    }
    if (status !== 'all') {
      where.status = { in: status.split(',') };
    }
    
    // 获取闲置资源
    const idleResources = await db.idleResource.findMany({
      where,
      include: {
        resource: {
          include: {
            cloudAccount: {
              select: { name: true, provider: true },
            },
          },
        },
      },
      orderBy: [
        { potentialSavings: 'desc' },
        { idleDays: 'desc' },
      ],
    });
    
    // 按闲置类型统计
    const typeStats = await db.idleResource.groupBy({
      by: ['idleType'],
      where,
      _count: { id: true },
      _sum: { potentialSavings: true, monthlyCost: true },
    });
    
    // 按状态统计
    const statusStats = await db.idleResource.groupBy({
      by: ['status'],
      _count: { id: true },
      _sum: { potentialSavings: true },
    });
    
    // 按资源类型统计
    const resourceTypeStats = await db.idleResource.findMany({
      where,
      select: {
        resource: { select: { type: true, category: true } },
        potentialSavings: true,
        monthlyCost: true,
      },
    });
    
    const typeCostMap = new Map<string, { count: number; savings: number; cost: number }>();
    for (const ir of resourceTypeStats) {
      const type = ir.resource.type;
      const existing = typeCostMap.get(type) || { count: 0, savings: 0, cost: 0 };
      typeCostMap.set(type, {
        count: existing.count + 1,
        savings: existing.savings + ir.potentialSavings,
        cost: existing.cost + ir.monthlyCost,
      });
    }
    
    // 总潜在节省
    const totalSavings = idleResources.reduce(
      (sum, ir) => sum + ir.potentialSavings,
      0
    );
    const totalMonthlyCost = idleResources.reduce(
      (sum, ir) => sum + ir.monthlyCost,
      0
    );
    
    // 格式化闲置资源数据
    const formattedResources = idleResources.map(ir => ({
      id: ir.id,
      resource: {
        id: ir.resource.id,
        name: ir.resource.name,
        type: ir.resource.type,
        category: ir.resource.category,
        account: ir.resource.cloudAccount.name,
        provider: ir.resource.cloudAccount.provider,
      },
      idleType: ir.idleType,
      avgCpuUsage: ir.avgCpuUsage,
      avgMemoryUsage: ir.avgMemoryUsage,
      idleDays: ir.idleDays,
      monthlyCost: ir.monthlyCost,
      potentialSavings: ir.potentialSavings,
      recommendation: ir.recommendation,
      status: ir.status,
      detectedAt: ir.detectedAt.toISOString(),
    }));
    
    // 闲置资源分布（按闲置天数分组）
    const idleDaysDistribution = [
      { range: '1-7天', count: 0, savings: 0 },
      { range: '8-14天', count: 0, savings: 0 },
      { range: '15-30天', count: 0, savings: 0 },
      { range: '30天以上', count: 0, savings: 0 },
    ];
    
    for (const ir of idleResources) {
      if (ir.idleDays <= 7) {
        idleDaysDistribution[0].count++;
        idleDaysDistribution[0].savings += ir.potentialSavings;
      } else if (ir.idleDays <= 14) {
        idleDaysDistribution[1].count++;
        idleDaysDistribution[1].savings += ir.potentialSavings;
      } else if (ir.idleDays <= 30) {
        idleDaysDistribution[2].count++;
        idleDaysDistribution[2].savings += ir.potentialSavings;
      } else {
        idleDaysDistribution[3].count++;
        idleDaysDistribution[3].savings += ir.potentialSavings;
      }
    }
    
    return NextResponse.json({
      resources: formattedResources,
      summary: {
        total: idleResources.length,
        totalSavings: Math.round(totalSavings * 100) / 100,
        totalMonthlyCost: Math.round(totalMonthlyCost * 100) / 100,
        byType: typeStats.map(t => ({
          idleType: t.idleType,
          count: t._count.id,
          savings: Math.round((t._sum.potentialSavings || 0) * 100) / 100,
          monthlyCost: Math.round((t._sum.monthlyCost || 0) * 100) / 100,
        })),
        byStatus: statusStats.map(s => ({
          status: s.status,
          count: s._count.id,
          savings: Math.round((s._sum.potentialSavings || 0) * 100) / 100,
        })),
        byResourceType: Array.from(typeCostMap.entries())
          .map(([type, data]) => ({
            type,
            count: data.count,
            savings: Math.round(data.savings * 100) / 100,
            cost: Math.round(data.cost * 100) / 100,
          }))
          .sort((a, b) => b.savings - a.savings),
        idleDaysDistribution: idleDaysDistribution.map(d => ({
          ...d,
          savings: Math.round(d.savings * 100) / 100,
        })),
      },
    });
  } catch (error) {
    console.error('获取闲置资源数据失败:', error);
    return NextResponse.json({ error: '获取数据失败' }, { status: 500 });
  }
}

// 更新闲置资源状态
export async function PATCH(request: NextRequest) {
  try {
    const body = await request.json();
    const { id, status } = body;
    
    const updated = await db.idleResource.update({
      where: { id },
      data: { status },
    });
    
    return NextResponse.json(updated);
  } catch (error) {
    console.error('更新闲置资源状态失败:', error);
    return NextResponse.json({ error: '更新失败' }, { status: 500 });
  }
}
