// 识别质量评测命令：读取清单，输出 FAR/FRR、EER、ROC 与活体指标。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"facial-liveness-verification-go/internal/adapter/onnx"
	"facial-liveness-verification-go/internal/eval"
)

func main() {
	var (
		manifestPath     = flag.String("manifest", "", "评测清单 YAML/JSON")
		modelDir         = flag.String("model-dir", "models", "ONNX 模型目录")
		outDir           = flag.String("out", "eval/out", "报告输出目录")
		targetFAR        = flag.Float64("target-far", 0.001, "目标误识率")
		livenessThresh   = flag.Float64("liveness-threshold", 0.85, "服务活体阈值")
		maxImpostorPairs = flag.Int("max-impostor-pairs", 0, "异人配对上限，0 表示全部")
		thresholdStep    = flag.Float64("threshold-step", 0, "阈值步长，0 表示使用清单配置")
		maxFRRAtTarget   = flag.Float64("max-frr-at-target", -1, "目标 FAR 下允许的最大 FRR，负数表示不设门禁")
		maxAPCER         = flag.Float64("max-apcer", -1, "允许的最大 APCER，负数表示不设门禁")
		maxBPCER         = flag.Float64("max-bpcer", -1, "允许的最大 BPCER，负数表示不设门禁")
		detectionThresh  = flag.Float64("detection-threshold", 0.9, "检测阈值")
		nmsThresh        = flag.Float64("nms-threshold", 0.3, "NMS IoU 阈值")
		analysisThresh   = flag.Float64("analysis-threshold", 0.6, "特征与活体路径的检测阈值")
		cropScale        = flag.Float64("liveness-crop-scale", 2.7, "活体裁剪倍率")
	)
	flag.Parse()
	if *manifestPath == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -manifest")
		os.Exit(2)
	}

	manifest, err := eval.LoadManifest(*manifestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "加载清单失败:", err)
		os.Exit(2)
	}
	if *maxImpostorPairs > 0 {
		manifest.Options.MaxImpostorPairs = *maxImpostorPairs
	}
	if *thresholdStep > 0 {
		manifest.Options.ThresholdStep = *thresholdStep
	}

	engine, err := onnx.New(onnx.Config{
		ModelDir:           *modelDir,
		DetectionThreshold: *detectionThresh,
		NMSThreshold:       *nmsThresh,
		AnalysisThreshold:  *analysisThresh,
		LivenessCropScale:  *cropScale,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "加载模型失败:", err)
		os.Exit(1)
	}
	defer engine.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runner := eval.NewRunner(engine, manifest)
	runner.TargetFAR = *targetFAR
	runner.LivenessThreshold = *livenessThresh
	report, pairs, samples, err := runner.Run(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "评测失败:", err)
		os.Exit(1)
	}
	if err := eval.WriteOutputs(*outDir, report, pairs, samples); err != nil {
		fmt.Fprintln(os.Stderr, "写报告失败:", err)
		os.Exit(1)
	}

	fmt.Printf("engine=%s\n", report.EngineVersion)
	fmt.Printf("identities=%d images=%d embedded=%d genuine_pairs=%d impostor_pairs=%d\n",
		report.Counts.Identities, report.Counts.Images, report.Counts.Embedded,
		report.Counts.GenuinePairs, report.Counts.ImpostorPairs)
	fmt.Printf("EER=%.6f at %.4f AUC=%.6f\n", report.EER, report.EERThreshold, report.AUC)
	fmt.Printf("target_far=%.6f threshold=%.4f FRR=%.6f achieved=%v\n",
		report.TargetFAR, report.RecommendedThreshold, report.FRRAtTarget, report.TargetAchieved)
	if report.Liveness != nil {
		point := report.Liveness.ServicePoint
		fmt.Printf("liveness threshold=%.4f APCER=%.6f BPCER=%.6f\n", point.Threshold, point.APCER, point.BPCER)
	}
	fmt.Printf("report=%s/report.json\n", *outDir)

	failed := false
	if !report.TargetAchieved {
		fmt.Fprintln(os.Stderr, "未达到目标 FAR")
		failed = true
	}
	if *maxFRRAtTarget >= 0 && report.FRRAtTarget > *maxFRRAtTarget {
		fmt.Fprintf(os.Stderr, "FRR %.6f 超过上限 %.6f\n", report.FRRAtTarget, *maxFRRAtTarget)
		failed = true
	}
	if report.Liveness != nil {
		point := report.Liveness.ServicePoint
		if *maxAPCER >= 0 && point.APCER > *maxAPCER {
			fmt.Fprintf(os.Stderr, "APCER %.6f 超过上限 %.6f\n", point.APCER, *maxAPCER)
			failed = true
		}
		if *maxBPCER >= 0 && point.BPCER > *maxBPCER {
			fmt.Fprintf(os.Stderr, "BPCER %.6f 超过上限 %.6f\n", point.BPCER, *maxBPCER)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}
