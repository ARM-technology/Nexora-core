package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type Info_AI struct {
	Name            string
	Age             string
	NameCompany     string
	MoreInfoCompany string
	MoreInfo        string
	// -----------------------
	SendMsg string
}

func AI_Employee(info Info_AI) (string, error) {

	ctx := context.Background()

	role := fmt.Sprintf(`
	Your Name :%s ,
	Your Age : %s  ,
	Your work / team in : %s ,
	Information about company : %s , 
	--------------------------------
You must act as this character: %s. 
	--------------------------------


Interaction Boundaries (Strictly Prohibited):

Your responses should be direct and brief, avoiding any long-windedness.

Politics: Any political discussion is strictly forbidden.

Inappropriate Content: Any content that violates ethics, morals, or public decency.

Sensitive Information:

Any details regarding the company's internal technologies.

Ongoing projects or the company's proprietary source code.

Any financial or strategic information related to the company.

You are an employee of this company. Do not discuss any topics outside the defined scope and do not provide any external information.
	`, info.Name, info.Age, info.NameCompany, info.MoreInfoCompany, info.MoreInfo)

	// تحميل ملف System.env
	err := godotenv.Load("System.env")
	if err != nil {
		log.Fatal("Error loading System.env file")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not found in System.env")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Printf("❌ فشل إنشاء العميل: %v", err)
		return "عذراً، النظام يواجه مشكلة تقنية.", nil
	}


	result, err := client.Models.GenerateContent(ctx, "gemini-2.5-pro",
		genai.Text(info.SendMsg),
		&genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(role, genai.RoleUser),
		})

	fmt.Println(result.Text())
	return result.Text(), nil
}
