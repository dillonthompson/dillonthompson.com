import type { NextPage } from 'next'
import Head from 'next/head'
import styles from '@/styles/Home.module.css'
import { Footer, Header } from '@/components/index'

const Home: NextPage = () => {
  return (
    <div className={styles.container}>
      <Head>
        <title>Dillon Thompson</title>
        <meta property="og:title" content="My page title" key="title" />
      </Head>
      <main>
        <Header />

        <Footer />
      </main>
    </div>
  )
}

export default Home
