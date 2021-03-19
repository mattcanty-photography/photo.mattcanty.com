package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type album struct {
	ID          string
	PortfolioID string
	CoverID     string
}

type photo struct {
	ID    string
	Album album
}

func getAlbums(portfolioID string) []album {
	svc := s3.New(session.New(), &aws.Config{Region: aws.String("eu-west-2")})

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(os.Getenv("PHOTO_BUCKET_NAME")),
		Prefix:    aws.String(fmt.Sprintf("portfolios/%s/albums/", portfolioID)),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		log.Fatal(err)
	}

	var albums []album

	for _, o := range resp.CommonPrefixes {
		albumID := strings.Split(*o.Prefix, "/")[3]
		coverID := getCoverPhotoID(portfolioID, albumID)

		albums = append(albums, album{
			ID:          albumID,
			PortfolioID: portfolioID,
			CoverID:     coverID,
		})
	}

	return albums
}

func getPhotos(portfolioID string, albumID string) []photo {
	svc := s3.New(session.New(), &aws.Config{Region: aws.String("eu-west-2")})

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(os.Getenv("PHOTO_BUCKET_NAME")),
		Prefix: aws.String(fmt.Sprintf("portfolios/%s/albums/%s/thumbs/", portfolioID, albumID)),
	})
	if err != nil {
		log.Fatal(err)
	}

	var photos []photo

	for _, o := range resp.Contents {
		keyElems := strings.Split(*o.Key, "/")
		photos = append(photos, photo{
			ID: keyElems[5],
			Album: album{
				ID:          keyElems[3],
				PortfolioID: keyElems[1],
			},
		})
	}

	return photos
}

func getCoverPhotoID(portfolioID string, albumID string) string {
	bucket := os.Getenv("PHOTO_BUCKET_NAME")
	prefix := fmt.Sprintf("portfolios/%s/albums/%s/thumbs/", portfolioID, albumID)

	svc := s3.New(session.New(), &aws.Config{Region: aws.String("eu-west-2")})

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int64(1),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Contents)

	coverID := strings.Split(*resp.Contents[0].Key, "/")[5]

	fmt.Println(coverID)

	return coverID
}
