package handlers

import (
	"chemistryuno/backend/repository"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ExportSubstancesToExcel 导出物质百科到Excel
func ExportSubstancesToExcel(c *gin.Context) {
	substanceRepo := repository.NewSubstanceRepository()

	// 获取所有已批准的物质（按组分组）
	substances, err := substanceRepo.FindApprovedGrouped()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询物质失败"})
		return
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Substances"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // 删除默认工作表

	// 设置表头
	headers := []string{"ID", "化学式", "物质名称", "元素组成", "状态", "创建者", "需要完善", "创建时间"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式（加粗）
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)

	// 填充数据
	for i, sub := range substances {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), sub.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), sub.Formula)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), sub.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), sub.Elements)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), sub.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), sub.CreatorName)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), sub.NeedsImprovement)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), sub.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 自动调整列宽
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	// 生成文件名
	filename := fmt.Sprintf("substances_%s.xlsx", time.Now().Format("20060102_150405"))

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")

	// 写入响应
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成Excel失败"})
		return
	}
}

// ExportReactionsToExcel 导出反应方程式到Excel
func ExportReactionsToExcel(c *gin.Context) {
	reactionRepo := repository.NewReactionRepository()

	// 获取所有已批准的反应（按组分组）
	reactions, err := reactionRepo.FindApprovedGrouped()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询反应失败"})
		return
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Reactions"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// 设置表头
	headers := []string{"ID", "反应物1", "反应物2", "反应方程式", "状态", "创建者", "创建时间"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)

	// 填充数据
	for i, rxn := range reactions {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), rxn.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), rxn.R1)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), rxn.R2)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), rxn.Display)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), rxn.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), rxn.CreatorName)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), rxn.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 自动调整列宽
	colWidths := map[string]float64{"A": 10, "B": 20, "C": 20, "D": 40, "E": 12, "F": 15, "G": 20}
	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	// 生成文件名
	filename := fmt.Sprintf("reactions_%s.xlsx", time.Now().Format("20060102_150405"))

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")

	// 写入响应
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成Excel失败"})
		return
	}
}

// ExportAllDataToExcel 导出物质和反应到同一个Excel文件（多工作表）
func ExportAllDataToExcel(c *gin.Context) {
	substanceRepo := repository.NewSubstanceRepository()
	reactionRepo := repository.NewReactionRepository()

	// 查询数据
	substances, err1 := substanceRepo.FindApprovedGrouped()
	reactions, err2 := reactionRepo.FindApprovedGrouped()

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询数据失败"})
		return
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// === 物质工作表 ===
	substanceSheet := "Substances"
	f.SetSheetName("Sheet1", substanceSheet)

	substanceHeaders := []string{"ID", "化学式", "物质名称", "元素组成", "状态", "创建者", "需要完善", "创建时间"}
	for i, header := range substanceHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(substanceSheet, cell, header)
	}

	// 表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(substanceSheet, "A1", fmt.Sprintf("%c1", 'A'+len(substanceHeaders)-1), headerStyle)

	for i, sub := range substances {
		row := i + 2
		f.SetCellValue(substanceSheet, fmt.Sprintf("A%d", row), sub.ID)
		f.SetCellValue(substanceSheet, fmt.Sprintf("B%d", row), sub.Formula)
		f.SetCellValue(substanceSheet, fmt.Sprintf("C%d", row), sub.Name)
		f.SetCellValue(substanceSheet, fmt.Sprintf("D%d", row), sub.Elements)
		f.SetCellValue(substanceSheet, fmt.Sprintf("E%d", row), sub.Status)
		f.SetCellValue(substanceSheet, fmt.Sprintf("F%d", row), sub.CreatorName)
		f.SetCellValue(substanceSheet, fmt.Sprintf("G%d", row), sub.NeedsImprovement)
		f.SetCellValue(substanceSheet, fmt.Sprintf("H%d", row), sub.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 自动调整列宽
	for i := 0; i < len(substanceHeaders); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(substanceSheet, col, col, 15)
	}

	// === 反应工作表 ===
	reactionSheet := "Reactions"
	f.NewSheet(reactionSheet)

	reactionHeaders := []string{"ID", "反应物1", "反应物2", "反应方程式", "状态", "创建者", "创建时间"}
	for i, header := range reactionHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(reactionSheet, cell, header)
	}

	f.SetCellStyle(reactionSheet, "A1", fmt.Sprintf("%c1", 'A'+len(reactionHeaders)-1), headerStyle)

	for i, rxn := range reactions {
		row := i + 2
		f.SetCellValue(reactionSheet, fmt.Sprintf("A%d", row), rxn.ID)
		f.SetCellValue(reactionSheet, fmt.Sprintf("B%d", row), rxn.R1)
		f.SetCellValue(reactionSheet, fmt.Sprintf("C%d", row), rxn.R2)
		f.SetCellValue(reactionSheet, fmt.Sprintf("D%d", row), rxn.Display)
		f.SetCellValue(reactionSheet, fmt.Sprintf("E%d", row), rxn.Status)
		f.SetCellValue(reactionSheet, fmt.Sprintf("F%d", row), rxn.CreatorName)
		f.SetCellValue(reactionSheet, fmt.Sprintf("G%d", row), rxn.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 调整列宽
	colWidths := map[string]float64{"A": 10, "B": 20, "C": 20, "D": 40, "E": 12, "F": 15, "G": 20}
	for col, width := range colWidths {
		f.SetColWidth(reactionSheet, col, col, width)
	}

	// 生成文件名
	filename := fmt.Sprintf("chemistry_data_%s.xlsx", time.Now().Format("20060102_150405"))

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")

	// 写入响应
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成Excel失败"})
		return
	}
}
