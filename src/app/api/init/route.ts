import { db } from '@/lib/db';
import { NextResponse } from 'next/server';

// 初始化模拟数据
export async function GET() {
  try {
    // 检查是否已有数据
    const existingAccounts = await db.cloudAccount.count();
    if (existingAccounts > 0) {
      return NextResponse.json({ message: '数据已存在，跳过初始化' });
    }

    // 创建云账户
    const accounts = await Promise.all([
      db.cloudAccount.create({
        data: {
          name: 'AWS Production',
          provider: 'aws',
          accountId: '123456789012',
          region: 'us-east-1',
          status: 'active',
          monthlyBudget: 50000,
        },
      }),
      db.cloudAccount.create({
        data: {
          name: 'GCP Development',
          provider: 'gcp',
          accountId: 'gcp-dev-001',
          region: 'us-central1',
          status: 'active',
          monthlyBudget: 20000,
        },
      }),
      db.cloudAccount.create({
        data: {
          name: 'Kubernetes Cluster',
          provider: 'kubernetes',
          accountId: 'k8s-prod-cluster',
          region: 'multi-region',
          status: 'active',
          monthlyBudget: 30000,
        },
      }),
    ]);

    // 为每个账户创建资源
    const resourceTypes = [
      { type: 'ec2', category: 'compute', costPerHour: 0.104 },
      { type: 's3', category: 'storage', costPerHour: 0.023 },
      { type: 'rds', category: 'database', costPerHour: 0.17 },
      { type: 'lambda', category: 'compute', costPerHour: 0.0000002083 },
      { type: 'pod', category: 'compute', costPerHour: 0.05 },
      { type: 'pvc', category: 'storage', costPerHour: 0.0001389 },
      { type: 'load-balancer', category: 'network', costPerHour: 0.0225 },
    ];

    const resources = [];
    for (const account of accounts) {
      // 每个账户创建10-20个资源
      const numResources = Math.floor(Math.random() * 10) + 10;
      for (let i = 0; i < numResources; i++) {
        const resType = resourceTypes[Math.floor(Math.random() * resourceTypes.length)];
        const resource = await db.resource.create({
          data: {
            cloudAccountId: account.id,
            resourceId: `${resType.type}-${account.provider}-${i + 1}`,
            name: `${resType.type}-${account.name.toLowerCase().replace(/\s/g, '-')}-${i + 1}`,
            type: resType.type,
            category: resType.category,
            region: account.region,
            status: Math.random() > 0.1 ? 'running' : 'stopped',
            specs: JSON.stringify({ cpu: '4', memory: '16GB', storage: '100GB' }),
            tags: JSON.stringify({ env: account.name.includes('Production') ? 'prod' : 'dev', team: 'backend' }),
            costPerHour: resType.costPerHour * (1 + Math.random() * 2),
          },
        });
        resources.push(resource);
      }
    }

    // 创建过去30天的成本记录
    const now = new Date();
    const categories = ['compute', 'storage', 'network', 'database', 'other'];
    const services: Record<string, string[]> = {
      compute: ['EC2', 'Lambda', 'Compute Engine', 'Kubernetes Pods'],
      storage: ['S3', 'EBS', 'Cloud Storage', 'PVC'],
      network: ['CloudFront', 'VPC', 'Load Balancer', 'NAT Gateway'],
      database: ['RDS', 'DynamoDB', 'Cloud SQL', 'MongoDB Atlas'],
      other: ['Support', 'Other Services'],
    };

    for (const account of accounts) {
      for (let day = 0; day < 30; day++) {
        const date = new Date(now);
        date.setDate(date.getDate() - day);
        
        for (const category of categories) {
          const serviceList = services[category];
          const service = serviceList[Math.floor(Math.random() * serviceList.length)];
          
          // 基础成本 + 随机波动
          let baseCost = 100 + Math.random() * 500;
          if (category === 'compute') baseCost *= 2;
          if (category === 'database') baseCost *= 1.5;
          
          // 添加趋势：近期成本略高
          const trendFactor = 1 + (30 - day) * 0.005;
          // 添加周周期性：周末成本较低
          const dayOfWeek = date.getDay();
          const weekFactor = (dayOfWeek === 0 || dayOfWeek === 6) ? 0.7 : 1;
          
          const finalCost = baseCost * trendFactor * weekFactor * (0.9 + Math.random() * 0.2);
          
          await db.costRecord.create({
            data: {
              cloudAccountId: account.id,
              date,
              category,
              service,
              cost: Math.round(finalCost * 100) / 100,
              currency: 'USD',
              usageQuantity: Math.round(finalCost * 10),
              usageUnit: category === 'storage' ? 'GB' : 'hours',
            },
          });
        }
      }
    }

    // 创建资源使用记录
    for (const resource of resources) {
      const numRecords = Math.floor(Math.random() * 24) + 24; // 24-48小时的数据
      for (let hour = 0; hour < numRecords; hour++) {
        const timestamp = new Date(now);
        timestamp.setHours(timestamp.getHours() - hour);
        
        // CPU使用率 - 正态分布
        let cpuUsage = Math.random() * 40 + 20; // 基础20-60%
        if (resource.category === 'compute') {
          cpuUsage = Math.random() * 60 + 10; // 计算资源波动更大
        }
        
        // 内存使用率 - 相对稳定
        const memoryUsage = 40 + Math.random() * 40;
        
        // 网络流量
        const networkIn = Math.random() * 1000;
        const networkOut = Math.random() * 500;
        
        // 磁盘使用
        const diskUsage = resource.category === 'storage' ? 50 + Math.random() * 40 : 20 + Math.random() * 30;
        
        await db.resourceUsage.create({
          data: {
            resourceId: resource.id,
            timestamp,
            cpuUsage: Math.min(100, cpuUsage),
            memoryUsage: Math.min(100, memoryUsage),
            networkIn,
            networkOut,
            diskUsage: Math.min(100, diskUsage),
            requestCount: Math.floor(Math.random() * 10000),
          },
        });
      }
    }

    // 创建浪费检测记录
    const wasteTypes = ['overprovisioned', 'unused', 'zombie', 'orphaned'];
    const severityLevels = ['low', 'medium', 'high', 'critical'];
    const recommendations = [
      '建议降低实例规格，当前规格过大',
      '资源已超过14天未使用，建议终止',
      '检测到孤立的存储卷，建议删除',
      '负载均衡器无后端实例，建议清理',
      '快照已过期，建议删除',
      '未关联的弹性IP，建议释放',
    ];

    for (const resource of resources.slice(0, Math.floor(resources.length * 0.3))) {
      // 30%的资源有浪费问题
      const wasteType = wasteTypes[Math.floor(Math.random() * wasteTypes.length)];
      const severity = severityLevels[Math.floor(Math.random() * severityLevels.length)];
      const estimatedSavings = resource.costPerHour * 24 * 30 * (0.3 + Math.random() * 0.5);
      
      await db.wasteDetection.create({
        data: {
          resourceId: resource.id,
          wasteType,
          severity,
          estimatedSavings: Math.round(estimatedSavings * 100) / 100,
          reason: `${wasteType}资源检测：CPU使用率长期低于10%`,
          recommendation: recommendations[Math.floor(Math.random() * recommendations.length)],
          status: Math.random() > 0.3 ? 'open' : 'acknowledged',
        },
      });
    }

    // 创建闲置资源记录
    const idleTypes = ['low_cpu', 'low_network', 'no_requests', 'stopped_long'];
    const idleRecommendations = ['terminate', 'downsize', 'schedule_stop'];

    for (const resource of resources.slice(0, Math.floor(resources.length * 0.2))) {
      // 20%的资源处于闲置状态
      const idleType = idleTypes[Math.floor(Math.random() * idleTypes.length)];
      const avgCpuUsage = Math.random() * 10; // 0-10%
      const avgMemoryUsage = 10 + Math.random() * 20; // 10-30%
      const idleDays = Math.floor(Math.random() * 20) + 5; // 5-25天
      const monthlyCost = resource.costPerHour * 24 * 30;
      
      await db.idleResource.create({
        data: {
          resourceId: resource.id,
          idleType,
          avgCpuUsage,
          avgMemoryUsage,
          idleDays,
          monthlyCost: Math.round(monthlyCost * 100) / 100,
          potentialSavings: Math.round(monthlyCost * 0.8 * 100) / 100,
          recommendation: idleRecommendations[Math.floor(Math.random() * idleRecommendations.length)],
          status: Math.random() > 0.5 ? 'active' : 'reviewing',
        },
      });
    }

    // 创建成本预测
    for (const account of accounts) {
      const predictionMonth = `${now.getFullYear()}-${String(now.getMonth() + 2).padStart(2, '0')}`;
      
      // 基于历史数据预测（简化版）
      const currentMonthSpend = (account.monthlyBudget || 30000) * (0.7 + Math.random() * 0.4);
      const predictedCost = currentMonthSpend * (1 + Math.random() * 0.2);
      
      await db.costPrediction.create({
        data: {
          cloudAccountId: account.id,
          predictionMonth,
          predictedCost: Math.round(predictedCost * 100) / 100,
          confidence: 0.75 + Math.random() * 0.2,
          trend: predictedCost > currentMonthSpend ? 'increasing' : predictedCost < currentMonthSpend * 0.95 ? 'decreasing' : 'stable',
          factors: JSON.stringify({
            historicalGrowth: '+5%',
            seasonalFactor: 1.1,
            plannedChanges: ['新增测试环境', '数据库扩容'],
          }),
        },
      });
    }

    // 创建预算告警
    for (const account of accounts) {
      const currentSpend = (account.monthlyBudget || 30000) * (0.6 + Math.random() * 0.3);
      
      await db.budgetAlert.create({
        data: {
          cloudAccountId: account.id,
          alertType: 'threshold',
          threshold: 80,
          currentSpend: Math.round(currentSpend * 100) / 100,
          budgetAmount: account.monthlyBudget || 30000,
          message: `本月支出已达到预算的 ${Math.round(currentSpend / (account.monthlyBudget || 30000) * 100)}%`,
          acknowledged: Math.random() > 0.5,
        },
      });
    }

    return NextResponse.json({
      message: '初始化成功',
      stats: {
        accounts: accounts.length,
        resources: resources.length,
        costRecords: '30天 x 账户数 x 分类',
      },
    });
  } catch (error) {
    console.error('初始化失败:', error);
    return NextResponse.json({ error: '初始化失败' }, { status: 500 });
  }
}
