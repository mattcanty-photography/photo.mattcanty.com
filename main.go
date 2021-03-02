package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/jfyne/live"
)

type portfolio struct {
	ID              string
	ThumbnailSource string
	AlternativeText string
	Title           string
	Description     string
}

type homePageInstance struct {
	Portfolios []portfolio
}

func main() {
	t, err := template.ParseFiles("root.html", "view.html")
	if err != nil {
		log.Fatal(err)
	}

	h, err := live.NewHandler(live.NewCookieStore("session-name", []byte("weak-secret")), live.WithTemplateRenderer(t))
	if err != nil {
		log.Fatal(err)
	}

	// Set the mount function for this handler.
	h.Mount = func(ctx context.Context, r *http.Request, s *live.Socket) (interface{}, error) {
		// This will initialise the chat for this socket.
		return newHomePageInstance(s), nil
	}

	http.Handle("/", h)
	http.Handle("/live.js", live.Javascript{})
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), nil)
}

func newHomePageInstance(s *live.Socket) *homePageInstance {
	m, ok := s.Assigns().(*homePageInstance)
	if !ok {
		return &homePageInstance{
			Portfolios: []portfolio{
				{
					ID:              live.NewID(),
					ThumbnailSource: "https://lh3.googleusercontent.com/pw/ACtC-3dDjSh_CYclTKNb9St0O4jgDPqkuE-4g46Fd3ZUlM1KO3qmmGsOj9xnvRNx6b0VNTh6IDDJTeNMaFUKHA-5-ySca8l1ZRjJQ7qGfZeSEblyZLv-zgghfynJHgACgcmwh84hJpat07P0K532EBuvzaxvaw=w2896-h1930-no?authuser=0",
					AlternativeText: "Microscope, Slide, Research, Close-Up, Test, Experiment",
					Title:           "Brockwell Park Community Greenhouses",
					Description:     "I volunteer weekly at Brockwell Park Community Greenhouses. Each time I go I spend 30 minutes capturing what has changed and anything which catches my eye.",
				},
				{
					ID:              live.NewID(),
					ThumbnailSource: "https://lh3.googleusercontent.com/QL_4fkTdRIFGz9KD6kWXuLvQ_3uWyJ6CdRfx_jG4eYkqaSJz3aSRUJT8sgoIekQ01JSuUoX5WXqhKV0Jx_jd4dWxkEZEjtz3hnultctRzJE2RrnsO3NF-RazMnNsvdzd-4s4JCW_v-KOO3gy516n03u_bxRXS32UoHiVE6_g2FHN6lFP3CP67fflUySlM-Ny-eSDGbdBnKHAVe_hVJMJZhAxk6FEu2xkKdiLUR4MzKI-PXEtN_FWtZCDjatxRuZ7KbRwism0je80EMdH03ljqfPKhpBRn2VeM7udsUH3OQgtHzV_KRiNCJn_yLpEGoqsAWXFQbSLLvOrrYLVEsHLH2tucmnhqYW7Avp0Q29MLHB2CZdvIw2zqd4u1m3FyTQcbeBuNieYzPwd0-dltH539w17dLykhrGNd1j8jvsf3RskeTONbJoFrET44EJYEKItKtFCKxBj9IVom9E10IU3dCIfxorgKzMB6WabnKdg3k3_uJVHWvoimuj4gFKdFFAiGm0q9p8CsxqlS69uWqp9_hEiMYxuseEGzzWdJgzI-S_Ybha365wABpod2dZKj7ovSROpEZWxOZ-q0aH8E90atxokj5pdW8RTe7h7kZ5lv3850lFqCICE11KMT1NIB4X5nDi35T6nxlCWlOGzspkERP2ZSvRrgJTnsbVMScC-vfyvnUZsWU_o7LzIEIGcOg=w1086-h724-no?authuser=0",
					AlternativeText: "Microscope, Slide, Research, Close-Up, Test, Experiment",
					Title:           "Cats",
					Description:     "Who doesn't love cats?",
				},
				{
					ID:              live.NewID(),
					ThumbnailSource: "https://lh3.googleusercontent.com/-lz71NvqJ32duBhABJBUcW4mt-8xXZGXWnrFdHCkfIkqHZt3-rbwINu0ynd0hc-htiFIk3f1nCAOCqlE_E5TogDsiP94rCi85DjAfSP5So1S3QmbcEnhfPeWwrJ7iwLVpPkCZA2u-9l4rfLA1VBGeuMrzgYrKSTbo8_B3KLkGd4oVCCvAyYFGS9kSg4tVpFRlXnvGwfTIp8p-VJ1nkD3fOCN2YE0fDuqFfx2Rcn_I9S5SrqERkcPx35f0pPiiy7D3qWgI3qJA2tZmgjWxV950Y5t8CFoNvlKLHaIGjTtJHa0m2WkjSgaqA_NRMZJ8GBzcCLMP_iwXdgDiCd_xNG9s66l2Uinwj4K3tYzRXsVNjYJOn1UIucxU4gIxxy5YC-Y9_9s4vytPw7kNCy3U4tMNHiu8owcIDUBxpAVddYxDD-KT8CdcX2bslvUMVi1JiXS0zJb5OnUobmJ5rmGGtw6i1ONa49vwWzmxMTGNhUIUwIBMQ7Eqh3uI1pyoKcAxkUWrzmierab7JgN2ldMezhppfTyXOCnRh6A-PWYn6DFH5dbekFx48yD0gYgdZM9nNtXs3586pWh0l05OoLWVm7FhbFXpQRYG929UcxvwBwMB3kPcQU05-NPb-kmC6VgMntE6upBHEu99ngIxtePbNEFgTqGAEqKPxlPBhyGCFLIEpgwo6FRWzYssKFNHU6IJA=s1930-no?authuser=0",
					AlternativeText: "Microscope, Slide, Research, Close-Up, Test, Experiment",
					Title:           "Doors",
					Description:     "Something my mum was interested in and she has passed it on to me. Nothing like a good door, especially when they're old.",
				},
			},
		}
	}
	return m
}
