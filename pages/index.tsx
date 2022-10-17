import type { NextPage } from 'next'
import Head from 'next/head'
import { Layout, HomeSlider } from '@/components/index'

const Home: NextPage = () => {
  return (
    <div>
      <Head>
        <title>Dillon Thompson</title>
        <meta property="og:title" content="Dillon Thompson" key="title" />
      </Head>
      <Layout>
        <HomeSlider />
        <h1>hello world</h1>
        <h2>hello world</h2>
        <h3>hello world</h3>
        <h4>hello world</h4>
        <p>hello world</p>
      </Layout>
    </div>
  )
}

export default Home
