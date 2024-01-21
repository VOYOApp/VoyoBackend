package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

var validate = validator.New()

func twilioSendOTP(phoneNumber string) (string, error) {
	// Create a new Twilio client using the account SID and auth token
	client := twilio.NewRestClient()

	// Create a new verification
	params := &verify.CreateVerificationParams{}
	params.SetTo(phoneNumber)
	params.SetChannel("sms")

	resp, err := client.VerifyV2.CreateVerification("VA6b5e4c3a246fb5fd51dc4594fd67b675", params)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return *resp.Status, nil
}

func twilioVerifyOTP(phoneNumber string, code string) (string, error) {
	// Create a new Twilio client using the account SID and auth token
	client := twilio.NewRestClient()

	// Create a new verification
	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(phoneNumber)
	params.SetCode(code)

	resp, err := client.VerifyV2.CreateVerificationCheck("VA6b5e4c3a246fb5fd51dc4594fd67b675", params)
	if err != nil {
		return "", err
	}

	return *resp.Status, nil
}

func sendOTP(c *fiber.Ctx) error {
	var body OTPData
	if err := c.BodyParser(&body); err != nil {
		log.Error("Error parsing the body:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := validate.Struct(body); err != nil {
		log.Error("Error validating the body:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	status, err := twilioSendOTP(body.PhoneNumber)
	if err != nil {
		log.Error("Error sending the OTP:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": fiber.StatusOK, "message status": status, "data": "OTP sent successfully"})
}

func verifyOTP(c *fiber.Ctx) error {
	var body VerifyData
	if err := c.BodyParser(&body); err != nil {
		log.Error("Error parsing the body:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := validate.Struct(body); err != nil {
		log.Error("Error validating the body:", err)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	status, err := twilioVerifyOTP(body.User.PhoneNumber, body.Code)
	if err != nil {
		log.Error("Error verifying the OTP:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": fiber.StatusOK, "message status": status, "data": "OTP verification successful"})
}
