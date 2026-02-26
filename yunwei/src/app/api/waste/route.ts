import { db } from '@/lib/db';
import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const severity = searchParams.get('severity');
    const wasteType = searchParams.get('wasteType');
    const status = searchParams.get('status') || 'open';
    
    // 构建查询条件
    const where: Record<string, unknown> = {};
    if (severity) {
      where.severity = severity;
    }
    if (wasteType) {
      where.wasteType = wasteType;
    }
    if (status !== 'all') {
      where.status = status;
    }
    
    // 获取浪费检测记录
    const wasteDetections = await db.wasteDetection.findMany({
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
        { severity: 'desc' },
        { estimatedSavings: 'desc' },
      ],
    });
    
    // 按严重程度统计
    const severityStats = await db.wasteDetection.groupBy({
      by: ['severity'],
      where,
      _count: { id: true },
      _sum: { estimatedSavings: true },
    });
    
    // 按浪费类型统计
    const typeStats = await db.wasteDetection.groupBy({
      by: ['wasteType'],
      where,
      _count: { id: true },
      _sum: { estimatedSavings: true },
    });
    
    // 按状态统计
    const statusStats = await db.wasteDetection.groupBy({
      by: ['status'],
      _count: { id: true },
      _sum: { estimatedSavings: true },
    });
    
    // 总潜在节省
    const totalSavings = wasteDetections.reduce(
      (sum, w) => sum + w.estimatedSavings,
      0
    );
    
    // 格式化浪费检测数据
    const formattedDetections = wasteDetections.map(w => ({
      id: w.id,
      resource: {
        id: w.resource.id,
        name: w.resource.name,
        type: w.resource.type,
        account: w.resource.cloudAccount.name,
        provider: w.resource.cloudAccount.provider,
      },
      wasteType: w.wasteType,
      severity: w.severity,
      estimatedSavings: w.estimatedSavings,
      reason: w.reason,
      recommendation: w.recommendation,
      status: w.status,
      detectedAt: w.detectedAt.toISOString(),
      resolvedAt: w.resolvedAt?.toISOString(),
    }));
    
    return NextResponse.json({
      detections: formattedDetections,
      summary: {
        total: wasteDetections.length,
        totalSavings: Math.round(totalSavings * 100) / 100,
        bySeverity: severityStats.map(s => ({
          severity: s.severity,
          count: s._count.id,
          savings: Math.round((s._sum.estimatedSavings || 0) * 100) / 100,
        })),
        byType: typeStats.map(t => ({
          wasteType: t.wasteType,
          count: t._count.id,
          savings: Math.round((t._sum.estimatedSavings || 0) * 100) / 100,
        })),
        byStatus: statusStats.map(s => ({
          status: s.status,
          count: s._count.id,
          savings: Math.round((s._sum.estimatedSavings || 0) * 100) / 100,
        })),
      },
    });
  } catch (error) {
    console.error('获取浪费检测数据失败:', error);
    return NextResponse.json({ error: '获取数据失败' }, { status: 500 });
  }
}

// 更新浪费检测状态
export async function PATCH(request: NextRequest) {
  try {
    const body = await request.json();
    const { id, status } = body;
    
    const updated = await db.wasteDetection.update({
      where: { id },
      data: {
        status,
        resolvedAt: status === 'resolved' ? new Date() : null,
      },
    });
    
    return NextResponse.json(updated);
  } catch (error) {
    console.error('更新浪费检测状态失败:', error);
    return NextResponse.json({ error: '更新失败' }, { status: 500 });
  }
}
