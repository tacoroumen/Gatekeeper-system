import cv2
import pytesseract
import subprocess

# Start the webcam feed
cap = cv2.VideoCapture(0)

while True:
    # Read the current frame
    _, image = cap.read()

    # Convert the image to gray scale
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)

    # Apply Gaussian blur
    blur = cv2.GaussianBlur(gray,(5,5),0)

    # Detect edges
    edges = cv2.Canny(blur, 50, 200)

    # Find contours
    contours, _ = cv2.findContours(edges, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)

    # Filter based on size
    plates = []
    for cnt in contours:
        x, y, w, h = cv2.boundingRect(cnt)
        if (w > 100 and w < 300) and (h > 50 and h < 120):
            plates.append((x, y, w, h))

    # Draw bounding boxes
    for (x, y, w, h) in plates:
        cv2.rectangle(image, (x, y), (x + w, y + h), (0, 255, 0), 2)

    # Use PyTesseract for OCR
    for (x, y, w, h) in plates:
        plate = gray[y:y+h, x:x+w]
        licenseplate = pytesseract.image_to_string(plate, config='-c tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 --psm 6')
        # Run terminal command with the plate as argument if it's exactly 6 characters long
        if len(licenseplate) == 7:
            #print('License Plate:', licenseplate)
            subprocess.run(["cmd", "/c", "go", "run", "main.go", "-plate", licenseplate])

    # Show the image
    cv2.imshow('Image', image)

    # Quit if 'q' is pressed
    if cv2.waitKey(1) & 0xFF == ord('q'):
        break

# Release the video capture object and close windows
cap.release()
cv2.destroyAllWindows()
