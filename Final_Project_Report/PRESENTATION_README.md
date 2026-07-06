# MCA Final Project Presentation

Generated file:

- `MCA_Final_Project_Presentation_Parag_Gulati.pptx`

## Slides Created

1. Title Slide
2. Project Introduction
3. Problem Statement
4. Objectives
5. Scope of Project
6. Technology Stack
7. System Architecture
8. Modules and Services Overview
9. Database Overview
10. API and Backend Flow
11. Frontend Overview
12. Step-by-Step User Guide: Access Frontend
13. Client/User Journey: Authentication and Discovery
14. Client/User Journey: Cart, Checkout and Payment
15. Client/User Journey: Profile and Notifications
16. Seller Dashboard Journey
17. Session Analytics Dashboard Journey
18. Superadmin Journey
19. Deployment and Local Run Overview
20. Testing and Validation
21. Challenges Faced
22. Limitations and Assumptions
23. Future Enhancements
24. Conclusion
25. Thank You / Q&A

## Screenshot Placeholders To Update

Headless Chrome and Edge both exited with code 13 in this environment, so screenshots could not be captured safely. The PPTX contains clear placeholders for:

- User App Home / Product Listing
- User App Login or Signup Page
- Search / Category Filter
- Product Detail
- Cart Page
- Checkout Page
- Profile / Orders / Notification Preferences
- Seller Product Manager / Orders / Analytics
- Analytics Overview / Funnel / Heatmap
- Superadmin Users / Payments / Audit Logs

Recommended local URLs after `docker compose up -d --build`:

- User App: `http://localhost:3000`
- Seller Dashboard: `http://localhost:3001`
- Session Analytics Dashboard: `http://localhost:3002`
- Superadmin Panel: `http://localhost:3003`
- API Gateway health: `http://localhost:8080/health/ready`

## Assumptions and Evidence

- Student name is assumed as `Parag Gulati` from the provided PDF/DOCX file names.
- The presentation content is based on the repository, `FINAL_PROJECT_REPORT.md`, `api/master-api.json`, frontend route files, database docs, Docker/Kubernetes files, CI workflow, and runbooks.
- No fake credentials, fake screenshots, fake production deployment claims, or unsupported modules were added.
- Buyer seed credentials were not found in `LOCAL_ACCESS_GUIDE.md`.
- Seller, session analytics admin, and superadmin demo credentials are documented in `LOCAL_ACCESS_GUIDE.md`; they are local/demo only and are not repeated as production credentials.
- Payment providers are disabled in local defaults, so the payment flow is described as implemented/configuration-dependent.
- Superadmin search is included only as a placeholder route because the frontend route is a protected placeholder module.
- Dedicated browser E2E automation was not found; unit/component/service tests and CI checks are documented.

## Suggested Final Update Before Submission

1. Run the platform locally using Docker Compose.
2. Capture the placeholder screenshots at 16:9 or full HD resolution.
3. Replace placeholder boxes in the PPTX with real screenshots.
4. Add college name, guide name, roll number, and submission date if required by your college format.