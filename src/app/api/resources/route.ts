import { db } from '@/lib/db';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const accountId = searchParams.get('accountId');
    const category = searchParams.get('category');
    const status = searchParams.get('status');
    const search = searchParams.get('search');
    
    // 构建查询条件
    const where: Record<string, unknown> = {};
    if (accountId) {
      where.cloudAccountId = accountId;
    }
    if (category) {
      where.category = category;
    }
    if (status) {
      where.status = status;
    }
    if (search) {
      where.OR = [
        { name: { contains: search } },
        { resourceId: { contains: search } },
        { type: { contains: search } },
      ];
    }
    
    // 获取资源列表
    const resources = await db.resource.findMany({
      where,
      include: {
        cloudAccount: {
          select: { name: true, provider: true },
        },
        usageRecords: {
          orderBy: { timestamp: 'desc' },
          take: 24, // 最近24小时的使用数据
        },
      },
      orderBy: { costPerHour: 'desc' },
    });
    
    // 计算每个资源的使用统计
    const resourcesWithStats = resources.map(r => {
      const usageRecords = r.usageRecords;
      let avgCpu = 0;
      let avgMemory = 0;
      let avgNetwork = 0;
      
      if (usageRecords.length > 0) {
        avgCpu = usageRecords.reduce((sum, u) => sum + u.cpuUsage, 0) / usageRecords.length;
        avgMemory = usageRecords.reduce((sum, u) => sum + u.memoryUsage, 0) / usageRecords.length;
        avgNetwork = usageRecords.reduce((sum, u) => sum + u.networkIn + u.networkOut, 0) / usageRecords.length;
      }
      
      // 月度成本估算
      const monthlyCost = r.costPerHour * 24 * 30;
      
      return {
        id: r.id,
        resourceId: r.resourceId,
        name: r.name,
        type: r.type,
        category: r.category,
        region: r.region,
        status: r.status,
        specs: r.specs ? JSON.parse(r.specs) : null,
        tags: r.tags ? JSON.parse(r.tags) : null,
        costPerHour: r.costPerHour,
        monthlyCost: Math.round(monthlyCost * 100) / 100,
        account: r.cloudAccount,
        stats: {
          avgCpu: Math.round(avgCpu * 10) / 10,
          avgMemory: Math.round(avgMemory * 10) / 10,
          avgNetwork: Math.round(avgNetwork * 10) / 10,
        },
        recentUsage: usageRecords.slice(0, 10).map(u => ({
          timestamp: u.timestamp,
          cpu: u.cpuUsage,
          memory: u.memoryUsage,
          network: u.networkIn + u.networkOut,
        })),
      };
    });
    
    // 按分类统计
    const categoryStats = await db.resource.groupBy({
      by: ['category'],
      where,
      _count: { id: true },
      _sum: { costPerHour: true },
    });
    
    // 按状态统计
    const statusStats = await db.resource.groupBy({
      by: ['status'],
      where,
      _count: { id: true },
    });
    
    // 按类型统计
    const typeStats = await db.resource.groupBy({
      by: ['type'],
      where,
      _count: { id: true },
      _sum: { costPerHour: true },
      orderBy: { _sum: { costPerHour: 'desc' } },
    });
    
    return NextResponse.json({
      resources: resourcesWithStats,
      stats: {
        total: resources.length,
        byCategory: categoryStats.map(c => ({
          category: c.category,
          count: c._count.id,
          hourlyCost: Math.round((c._sum.costPerHour || 0) * 1000) / 1000,
          monthlyCost: Math.round((c._sum.costPerHour || 0) * 24 * 30 * 100) / 100,
        })),
        byStatus: statusStats.map(s => ({
          status: s.status,
          count: s._count.id,
        })),
        byType: typeStats.slice(0, 10).map(t => ({
          type: t.type,
          count: t._count.id,
          hourlyCost: Math.round((t._sum.costPerHour || 0) * 1000) / 1000,
        })),
      },
    });
  } catch (error) {
    console.error('获取资源数据失败:', error);
    return NextResponse.json({ error: '获取数据失败' }, { status: 500 });
  }
}
