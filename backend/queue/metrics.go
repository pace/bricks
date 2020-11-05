package queue

import (
	"context"
	"time"

	"github.com/adjust/rmq/v3"
	pberrors "github.com/pace/bricks/maintenance/errors"
	"github.com/pace/bricks/maintenance/log"
	"github.com/pace/bricks/pkg/routine"
	"github.com/prometheus/client_golang/prometheus"
)

type queueStatsGauges struct {
	readyGauge      *prometheus.GaugeVec
	rejectedGauge   *prometheus.GaugeVec
	connectionGauge *prometheus.GaugeVec
	consumerGauge   *prometheus.GaugeVec
	unackedGauge    *prometheus.GaugeVec
}

func gatherMetrics(connection rmq.Connection) {
	gauges := registerConnection(connection)
	ctx := log.ContextWithSink(log.WithContext(context.Background()), new(log.Sink))

	routine.Run(ctx, func(_ context.Context) {
		queues, err := connection.GetOpenQueues()
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("rmq metrics: could not get open queues")
			pberrors.Handle(ctx, err)
		}
		stats, err := connection.CollectStats(queues)
		if err != nil {
			log.Ctx(ctx).Debug().Err(err).Msg("rmq metrics: could not collect stats")
			pberrors.Handle(ctx, err)
		}
		for queue, queueStats := range stats.QueueStats {
			labels := prometheus.Labels{
				"queue": queue,
			}
			gauges.readyGauge.With(labels).Set(float64(queueStats.ReadyCount))
			gauges.rejectedGauge.With(labels).Set(float64(queueStats.RejectedCount))
			gauges.connectionGauge.With(labels).Set(float64(queueStats.ConnectionCount()))
			gauges.consumerGauge.With(labels).Set(float64(queueStats.ConsumerCount()))
			gauges.unackedGauge.With(labels).Set(float64(queueStats.UnackedCount()))
		}
		time.Sleep(cfg.MetricsRefreshInterval)
	})
}

func registerConnection(connection rmq.Connection) queueStatsGauges {
	gauges := queueStatsGauges{
		readyGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "rmq",
			Name:      "ready",
			Help:      "Number of ready messages on queue",
		}, []string{"queue"}),
		rejectedGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "rmq",
			Name:      "rejected",
			Help:      "Number of rejected messages on queue",
		}, []string{"queue"}),
		connectionGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "rmq",
			Name:      "connection",
			Help:      "Number of connections consuming a queue",
		}, []string{"queue"}),
		consumerGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "rmq",
			Name:      "consumer",
			Help:      "Number of consumers consuming messages for a queue",
		}, []string{"queue"}),
		unackedGauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "rmq",
			Name:      "unacked",
			Help:      "Number of unacked messages on a consumer",
		}, []string{"queue"}),
	}

	prometheus.MustRegister(gauges.readyGauge)
	prometheus.MustRegister(gauges.rejectedGauge)
	prometheus.MustRegister(gauges.connectionGauge)
	prometheus.MustRegister(gauges.consumerGauge)
	prometheus.MustRegister(gauges.unackedGauge)

	return gauges
}
