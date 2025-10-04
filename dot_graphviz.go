package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-graphviz"
)

func runDotToImage(outfname string, format string, dot []byte) (string, error) {
    // 首先尝试使用 go-graphviz
    ctx := context.Background()
    
    // 设置超时以避免无限等待
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    g, err := graphviz.New(ctx)
    if err != nil {
        log.Printf("内置渲染器初始化失败，切换到系统 Graphviz...")
        if *debugFlag {
            log.Printf("go-graphviz initialization failed: %v, falling back to system graphviz", err)
        }
        return runDotToImageCallSystemGraphviz(outfname, format, dot)
    }
    
    graph, err := graphviz.ParseBytes(dot)
    if err != nil {
        log.Printf("内置渲染器解析失败，切换到系统 Graphviz...")
        if *debugFlag {
            log.Printf("go-graphviz parse failed: %v, falling back to system graphviz", err)
        }
        g.Close()
        return runDotToImageCallSystemGraphviz(outfname, format, dot)
    }
    
    var img string
    if outfname == "" {
        img = filepath.Join(os.TempDir(), fmt.Sprintf("go-callvis_export.%s", format))
    } else {
        // 检查是否已经有正确的扩展名
        expectedExt := "." + format
        if filepath.Ext(outfname) == expectedExt {
            img = outfname
        } else {
            img = fmt.Sprintf("%s.%s", outfname, format)
        }
        
        // 确保输出目录存在
        if dir := filepath.Dir(img); dir != "." {
            if err := os.MkdirAll(dir, 0755); err != nil {
                log.Printf("创建输出目录失败，切换到系统 Graphviz...")
                if *debugFlag {
                    log.Printf("failed to create output directory: %v, falling back to system graphviz", err)
                }
                graph.Close()
                g.Close()
                return runDotToImageCallSystemGraphviz(outfname, format, dot)
            }
        }
    }
    
    // 尝试使用 bytes.Buffer 和 Render 方法
    var buf bytes.Buffer
    renderErr := g.Render(ctx, graph, graphviz.Format(format), &buf)
    
    // 清理资源
    if closeErr := graph.Close(); closeErr != nil {
        // 只在调试模式下显示详细错误
        if *debugFlag {
            log.Printf("error closing graph: %v", closeErr)
        }
    }
    if closeErr := g.Close(); closeErr != nil {
        // 只在调试模式下显示详细错误
        if *debugFlag {
            log.Printf("error closing graphviz: %v", closeErr)
        }
    }
    
    if renderErr != nil {
        log.Printf("使用内置渲染器失败，切换到系统 Graphviz...")
        if *debugFlag {
            log.Printf("go-graphviz render failed: %v, falling back to system graphviz", renderErr)
        }
        return runDotToImageCallSystemGraphviz(outfname, format, dot)
    }
    
    // 手动写入文件
    if writeErr := os.WriteFile(img, buf.Bytes(), 0644); writeErr != nil {
        log.Printf("写入文件失败，切换到系统 Graphviz...")
        if *debugFlag {
            log.Printf("failed to write output file: %v, falling back to system graphviz", writeErr)
        }
        return runDotToImageCallSystemGraphviz(outfname, format, dot)
    }
    
    return img, nil
}
